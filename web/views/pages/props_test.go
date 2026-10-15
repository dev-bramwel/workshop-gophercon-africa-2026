package pages_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"
	"uuid"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
	"workshop/web/views/pages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtures() []sighting.Sighting {
	result := make([]sighting.Sighting, 30)
	for i := range result {
		result[i] = sighting.Sighting{ID: uuid.New(), HappenedAt: time.Date(2024, 1, i+1, 12, 0, 0, 0, time.UTC), Species: species.Species{GBIFKey: 5219404, ScientificName: "Panthera leo", VernacularName: "lion"}, Location: sighting.Location{Country: "Kenya", CountryCode: country.Kenya}}
		if i >= 20 {
			result[i].Species = species.Species{GBIFKey: 2435270, ScientificName: "Acinonyx jubatus", VernacularName: "cheetah"}
			result[i].Location = sighting.Location{Country: "South Africa", CountryCode: country.SouthAfrica}
		}
	}
	return result
}
func TestCollectionFiltersAndReports(t *testing.T) {
	t.Parallel()
	items := fixtures()
	d := pages.NewCollection(items, "/sightings", "5219404", "KE", "lion", "newest", 2)
	assert.Len(t, d.Items, 20)
	assert.Len(t, d.Visible, 8)
	assert.Equal(t, 20, d.Summary.Sightings)
	assert.Equal(t, 1, d.Summary.Species)
	assert.Equal(t, 1, d.Summary.Countries)
	assert.Equal(t, 2, d.Pages)
	assert.Equal(t, 2, d.Page)
	assert.Equal(t, 20, d.ByAnimal[0].Count)
	assert.Equal(t, 20, d.Species[0].Count)
	assert.Len(t, d.Animals, 2, "options remain available after filtering")
	assert.Len(t, d.Countries, 2)
	require.NotEmpty(t, d.Visible)
	assert.True(t, d.Items[0].HappenedAt.After(d.Items[1].HappenedAt))
	assert.Equal(t, 30, len(items), "building a view must not mutate the source slice")
	u, err := url.Parse(d.Link("/map", 3))
	require.NoError(t, err)
	assert.Equal(t, "5219404", u.Query().Get("animal"))
	assert.Equal(t, "KE", u.Query().Get("country"))
	assert.Equal(t, "lion", u.Query().Get("q"))
	assert.Equal(t, "3", u.Query().Get("page"))
}
func TestCollectionEmptyAndPageBounds(t *testing.T) {
	t.Parallel()
	for _, page := range []int{-3, 0, 1, 999} {
		t.Run(fmt.Sprint(page), func(t *testing.T) {
			d := pages.NewCollection(fixtures(), "/sightings", "", "", "no such animal", "", page)
			assert.Empty(t, d.Visible)
			assert.Empty(t, d.Items)
			assert.Equal(t, 1, d.Page)
			assert.Equal(t, 1, d.Pages)
			assert.Zero(t, d.Summary.Sightings)
		})
	}
	d := pages.NewCollection(fixtures(), "/sightings", "", "", "", "oldest", 999)
	assert.Equal(t, 3, d.Page)
	assert.Len(t, d.Visible, 6)
	assert.True(t, d.Items[0].HappenedAt.Before(d.Items[1].HappenedAt))
}
func TestAnimalRepresentativeAndPhotoCounts(t *testing.T) {
	t.Parallel()
	items := fixtures()[:2]
	items[0].Media = []sighting.Media{{Type: "Sound", URL: "https://example.org/sound.mp3"}}
	items[1].Media = []sighting.Media{{Type: "StillImage", URL: "https://example.org/lion.jpg"}, {Type: "StillImage", URL: "javascript:bad()"}}
	d := pages.NewCollection(items, "/gallery", "", "", "", "oldest", 1)
	require.Len(t, d.Species, 1)
	assert.Equal(t, items[1].ID, d.Species[0].Sighting.ID, "prefer a sighting that has a usable photo")
	assert.Equal(t, 1, d.Photos)
	assert.Equal(t, "https://example.org/lion.jpg", pages.Photo(d.Species[0].Sighting))
	assert.Equal(t, "Weather unavailable", pages.WeatherText(items[0]))
}

func TestHeroSlidesRotateDistinctPhotographedAnimals(t *testing.T) {
	t.Parallel()
	items := fixtures()[:4]
	for i := range items {
		items[i].Media = []sighting.Media{{Type: "StillImage", URL: fmt.Sprintf("https://example.org/photo-%d.jpg", i)}}
	}
	items[2].Species = species.Species{GBIFKey: 2435350, VernacularName: "elephant"}
	items[3].Species = species.Species{GBIFKey: 2435270, VernacularName: "cheetah"}
	items[3].Media = nil
	d := pages.NewCollection(items, "/", "", "", "", "newest", 1)
	slides := d.HeroSlides()
	require.Len(t, slides, 2, "duplicate animals and sightings without photos are excluded")
	assert.Equal(t, d.Hero.ID, slides[0].ID, "the existing highlight is the first slide")
	assert.Equal(t, 2435350, slides[0].Species.GBIFKey)
	assert.Equal(t, 5219404, slides[1].Species.GBIFKey)
	assert.Empty(t, (pages.Collection{}).HeroSlides())
	single := pages.NewCollection(items[:1], "/", "", "", "", "", 1)
	assert.Len(t, single.HeroSlides(), 1)
}
