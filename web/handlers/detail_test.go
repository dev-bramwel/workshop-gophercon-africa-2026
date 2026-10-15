package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
	"workshop/internal/domain/weather"
	"workshop/web/handlers"
	"workshop/web/views/pages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFinder struct {
	item sighting.Sighting
	err  error
}

func (f fakeFinder) Find(context.Context, uuid.UUID) (sighting.Sighting, error) { return f.item, f.err }
func testLogger() *slog.Logger                                                  { return slog.New(slog.NewTextHandler(io.Discard, nil)) }
func sampleSighting() sighting.Sighting {
	return sighting.Sighting{ID: uuid.New(), GBIFKey: 123, Species: species.Species{GBIFKey: 5219404, VernacularName: "lion", ScientificName: "Panthera leo", Family: "Felidae", Genus: "Panthera"}, Location: sighting.Location{Country: "Kenya", CountryCode: country.Kenya}, Weather: weather.Weather{ObservedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Temperature: weather.Temperature{Actual: 24.5}, Condition: weather.Condition{Description: "Clear sky"}}}
}
func TestSightingDetail(t *testing.T) {
	t.Parallel()
	s := sampleSighting()
	for _, tt := range []struct {
		name, id string
		finder   fakeFinder
		status   int
	}{
		{"valid", s.ID.String(), fakeFinder{item: s}, 200},
		{"malformed", "not-a-uuid", fakeFinder{item: s}, 404},
		{"missing", uuid.New().String(), fakeFinder{err: sighting.ErrNotFound}, 404},
		{"store error", s.ID.String(), fakeFinder{err: errors.New("private database error")}, 500},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/sightings/"+tt.id, nil)
			r.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			handlers.Sighting(testLogger(), tt.finder)(w, r)
			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
			assert.NotContains(t, w.Body.String(), "private database error")
			if tt.status == 200 {
				assert.Contains(t, w.Body.String(), "Clear sky")
				assert.Contains(t, w.Body.String(), "Historical conditions")
				assert.Contains(t, w.Body.String(), "/animals/5219404")
			}
		})
	}
}
func TestCollectionPages(t *testing.T) {
	t.Parallel()
	s := sampleSighting()
	for _, path := range []string{"/", "/sightings", "/gallery", "/map", "/statistics"} {
		t.Run(path, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, path+"?animal=5219404&country=ke", nil)
			w := httptest.NewRecorder()
			handlers.Collection(testLogger(), fakeLister{items: []sighting.Sighting{s}}, path)(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Wildlens")
			assert.NotContains(t, w.Body.String(), "for _,")
			assert.NotContains(t, w.Body.String(), "@components")
			assert.NotContains(t, w.Body.String(), "WeatherText(s)")
			assert.NotContains(t, w.Body.String(), "s.Location.Country")
			if path == "/sightings" || path == "/map" || path == "/" {
				assert.Contains(t, w.Body.String(), "Clear sky", "cards show the stored weather")
				assert.Contains(t, w.Body.String(), "Kenya")
			}
		})
	}
	for _, query := range []string{"animal=bad", "animal=-1", "country=Kenya", "sort=bad"} {
		t.Run(query, func(t *testing.T) {
			w := httptest.NewRecorder()
			handlers.Collection(testLogger(), fakeLister{}, "/sightings")(w, httptest.NewRequest(http.MethodGet, "/sightings?"+query, nil))
			assert.Equal(t, 400, w.Code)
		})
	}
	w := httptest.NewRecorder()
	handlers.Home(testLogger(), fakeLister{})(w, httptest.NewRequest(http.MethodGet, "/unknown/path", nil))
	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "Even explorers get lost")
}
func TestAnimalDetail(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		id     string
		status int
	}{{"5219404", 200}, {"not-an-animal", 404}, {"-1", 404}, {"999999", 404}} {
		t.Run(tt.id, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/animals/"+tt.id, nil)
			r.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()
			handlers.Animal(testLogger(), fakeLister{items: []sighting.Sighting{sampleSighting()}})(w, r)
			assert.Equal(t, tt.status, w.Code)
			if tt.status == 200 {
				assert.Contains(t, w.Body.String(), "Did you know?")
				assert.Contains(t, w.Body.String(), "San Diego Zoo")
				assert.Contains(t, w.Body.String(), "sightings?animal=5219404")
			}
		})
	}
}
func TestObservationTextIsEscaped(t *testing.T) {
	t.Parallel()
	s := sampleSighting()
	s.RecordedBy = "<script>alert('observer')</script>"
	s.Remarks = "<img src=x onerror=alert(1)>"
	var b bytes.Buffer
	require.NoError(t, pages.SightingDetail(s).Render(t.Context(), &b))
	assert.NotContains(t, b.String(), s.RecordedBy)
	assert.NotContains(t, b.String(), s.Remarks)
	assert.Contains(t, b.String(), "&lt;script&gt;")
	assert.False(t, strings.Contains(b.String(), "javascript:"))
}
