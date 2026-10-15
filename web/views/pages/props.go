package pages

import (
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/species"
)

// Collection contains one consistent snapshot of the filtered records and their report.
type Collection struct {
	Path, Animal, Country, Search, Sort string
	Items, Visible                      []sighting.Sighting
	Animals, Countries                  []Option
	Summary                             sighting.Summary
	Page, Pages, Photos                 int
	Species                             []AnimalCard
	Hero                                sighting.Sighting
	ByAnimal, ByCountry, ByMonth        []sighting.Tally
}
type Option struct{ Value, Label string }
type AnimalCard struct {
	Species  species.Species
	Sighting sighting.Sighting
	Count    int
}
type Fact struct{ Title, Text, Source, SourceName string }

func NewCollection(items []sighting.Sighting, path, animal, country, search, sort string, page int) Collection {
	d := Collection{Path: path, Animal: animal, Country: country, Search: search, Sort: sort}
	animals, countries := map[string]string{}, map[string]string{}
	for _, s := range items {
		animals[strconv.Itoa(s.Species.GBIFKey)] = Name(s.Species)
		if s.Location.CountryCode != "" {
			countries[s.Location.CountryCode.String()] = s.Location.Country
		}
		if animal != "" && strconv.Itoa(s.Species.GBIFKey) != animal {
			continue
		}
		if country != "" && s.Location.CountryCode.String() != country {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(Name(s.Species)+" "+s.Species.ScientificName+" "+Place(s.Location)+" "+s.RecordedBy), strings.ToLower(search)) {
			continue
		}
		d.Items = append(d.Items, s)
	}
	for k, v := range animals {
		d.Animals = append(d.Animals, Option{k, v})
	}
	for k, v := range countries {
		d.Countries = append(d.Countries, Option{k, v})
	}
	slices.SortFunc(d.Animals, func(a, b Option) int { return strings.Compare(a.Label, b.Label) })
	slices.SortFunc(d.Countries, func(a, b Option) int { return strings.Compare(a.Label, b.Label) })
	slices.SortStableFunc(d.Items, func(a, b sighting.Sighting) int {
		if sort == "oldest" {
			return a.HappenedAt.Compare(b.HappenedAt)
		}
		if sort == "animal" {
			return strings.Compare(Name(a.Species), Name(b.Species))
		}
		return b.HappenedAt.Compare(a.HappenedAt)
	})
	d.Summary = sighting.Summarise(d.Items)
	grouped := map[int]int{}
	for _, s := range d.Items {
		for _, m := range s.Media {
			if IsPhoto(m) {
				d.Photos++
			}
		}
		if i, ok := grouped[s.Species.GBIFKey]; ok {
			d.Species[i].Count++
			if Photo(d.Species[i].Sighting) == "" && Photo(s) != "" {
				d.Species[i].Sighting = s
			}
		} else {
			grouped[s.Species.GBIFKey] = len(d.Species)
			d.Species = append(d.Species, AnimalCard{s.Species, s, 1})
		}
	}
	slices.SortStableFunc(d.Species, func(a, b AnimalCard) int { return cmp.Compare(b.Count, a.Count) })
	if len(d.Items) > 0 {
		d.Hero = d.Items[0]
		for _, s := range d.Items {
			if Photo(s) != "" {
				d.Hero = s
				break
			}
		}
		for _, s := range d.Items {
			if slices.Contains([]int{2435350, 2435349, 5219461}, s.Species.GBIFKey) && Photo(s) != "" {
				d.Hero = s
				break
			}
		}
	}
	d.ByAnimal = sighting.Top(sighting.CountBy(d.Items, func(s sighting.Sighting) string { return Name(s.Species) }), 7)
	d.ByCountry = sighting.Top(sighting.CountBy(d.Items, func(s sighting.Sighting) string { return s.Location.Country }), 7)
	months := sighting.CountBy(d.Items, func(s sighting.Sighting) string {
		if s.HappenedAt.IsZero() {
			return ""
		}
		return s.HappenedAt.Format("2006-01")
	})
	for k, v := range months {
		d.ByMonth = append(d.ByMonth, sighting.Tally{Name: k, Count: v})
	}
	slices.SortFunc(d.ByMonth, func(a, b sighting.Tally) int { return strings.Compare(a.Name, b.Name) })
	if len(d.ByMonth) > 12 {
		d.ByMonth = d.ByMonth[len(d.ByMonth)-12:]
	}
	d.Pages = max(1, (len(d.Items)+11)/12)
	d.Page = min(max(1, page), d.Pages)
	start := (d.Page - 1) * 12
	d.Visible = d.Items[start:min(start+12, len(d.Items))]
	return d
}
func (d Collection) Link(path string, page int) string {
	q := url.Values{}
	if d.Animal != "" {
		q.Set("animal", d.Animal)
	}
	if d.Country != "" {
		q.Set("country", d.Country)
	}
	if d.Search != "" {
		q.Set("q", d.Search)
	}
	if d.Sort != "" {
		q.Set("sort", d.Sort)
	}
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) > 0 {
		return path + "?" + q.Encode()
	}
	return path
}

// HeroSlides rotates through one photographed sighting per animal, starting
// with the existing highlighted encounter.
func (d Collection) HeroSlides() []sighting.Sighting {
	slides := make([]sighting.Sighting, 0, len(d.Species))
	seen := make(map[int]bool)
	add := func(s sighting.Sighting) {
		if Photo(s) == "" || seen[s.Species.GBIFKey] {
			return
		}
		seen[s.Species.GBIFKey] = true
		slides = append(slides, s)
	}
	add(d.Hero)
	for _, a := range d.Species {
		add(a.Sighting)
	}
	return slides
}

func (d Collection) Featured() []AnimalCard      { return d.Species[:min(4, len(d.Species))] }
func (d Collection) Recent() []sighting.Sighting { return d.Items[:min(4, len(d.Items))] }
func Name(s species.Species) string {
	n := cmp.Or(s.VernacularName, s.CanonicalName, s.ScientificName, "Unknown species")
	runes := []rune(n)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
func Scientific(s species.Species) string {
	return cmp.Or(s.CanonicalName, s.ScientificName, "Taxonomy not recorded")
}
func Place(l sighting.Location) string {
	p := []string{}
	for _, v := range []string{l.Locality, l.StateProvince, l.Country} {
		if v != "" {
			p = append(p, v)
		}
	}
	return cmp.Or(strings.Join(p, ", "), "Location not recorded")
}
func Date(s sighting.Sighting) string {
	if s.HappenedAt.IsZero() {
		return "Date not recorded"
	}
	return s.HappenedAt.Format("02 Jan 2006")
}
func IsPhoto(m sighting.Media) bool {
	return (m.Type == "StillImage" || strings.HasPrefix(m.Format, "image/")) && (strings.HasPrefix(m.URL, "https://") || strings.HasPrefix(m.URL, "http://"))
}
func Photo(s sighting.Sighting) string {
	for _, m := range s.Media {
		if IsPhoto(m) {
			u := m.URL
			if strings.Contains(u, "inaturalist") {
				u = strings.Replace(u, "/original.", "/large.", 1)
			}
			return u
		}
	}
	return ""
}
func SightingURL(s sighting.Sighting) string { return "/sightings/" + s.ID.String() }
func AnimalURL(s species.Species) string     { return "/animals/" + strconv.Itoa(s.GBIFKey) }
func Number(n int) string {
	raw := strconv.Itoa(n)
	for i := len(raw) - 3; i > 0; i -= 3 {
		raw = raw[:i] + "," + raw[i:]
	}
	return raw
}
func WeatherText(s sighting.Sighting) string {
	if s.Weather.ObservedAt.IsZero() {
		return "Weather unavailable"
	}
	return fmt.Sprintf("%.0f°C · %s", s.Weather.Temperature.Actual, cmp.Or(s.Weather.Condition.Description, "Conditions recorded"))
}
func Percent(n, total int) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", float64(n)*100/float64(total))
}
func BarStyle(n int, rows []sighting.Tally) string {
	peak := 1
	for _, r := range rows {
		peak = max(peak, r.Count)
	}
	return fmt.Sprintf("--bar: %.2f%%", float64(n)*100/float64(peak))
}
func GBIF(s species.Species) string { return "https://www.gbif.org/species/" + strconv.Itoa(s.GBIFKey) }

func CoverMedia(s sighting.Sighting) sighting.Media {
	for _, m := range s.Media {
		if IsPhoto(m) {
			return m
		}
	}
	return sighting.Media{}
}
func PhotoCredit(s sighting.Sighting) string {
	m := CoverMedia(s)
	return cmp.Or(m.Creator, m.RightsHolder, "Photographer not recorded")
}
func LicenseName(raw string) string {
	if strings.Contains(raw, "creativecommons.org/licenses/") {
		parts := strings.Split(strings.Trim(raw, "/"), "/")
		if len(parts) >= 2 {
			return "CC " + strings.ToUpper(parts[len(parts)-2]) + " " + parts[len(parts)-1]
		}
	}
	if strings.Contains(raw, "creativecommons.org/publicdomain/zero/") {
		return "CC0"
	}
	if raw != "" {
		return "See licence"
	}
	return "Licence not recorded"
}
