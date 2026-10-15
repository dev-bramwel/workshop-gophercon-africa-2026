package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/web/views/pages"
)

type SightingsLister interface {
	List(context.Context, sighting.ListFilter) ([]sighting.Sighting, error)
}

func Home(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return Collection(logger, sightings, "/")
}

func Collection(logger *slog.Logger, sightings SightingsLister, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			renderError(r.Context(), w, logger, http.StatusNotFound, "This trail has gone a little wild.")
			return
		}
		q := r.URL.Query()
		animal := q.Get("animal")
		if animal != "" {
			if n, err := strconv.Atoi(animal); err != nil || n <= 0 {
				renderError(r.Context(), w, logger, 400, "Choose an animal from the available filters.")
				return
			}
		}
		code, err := country.Parse(q.Get("country"))
		if err != nil {
			renderError(r.Context(), w, logger, 400, "Choose a country from the available filters.")
			return
		}
		sort := q.Get("sort")
		if sort != "" && sort != "newest" && sort != "oldest" && sort != "animal" {
			renderError(r.Context(), w, logger, 400, "Choose a valid sort order.")
			return
		}
		items, err := sightings.List(r.Context(), sighting.ListFilter{})
		if err != nil {
			logger.ErrorContext(r.Context(), "handlers: listing sightings", "error", err)
			renderError(r.Context(), w, logger, 500, "Unable to load sightings. Please try again shortly.")
			return
		}
		page, _ := strconv.Atoi(q.Get("page"))
		d := pages.NewCollection(items, path, animal, code.String(), strings.TrimSpace(q.Get("q")), sort, page)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pages.CollectionPage(d).Render(r.Context(), w); err != nil {
			logger.ErrorContext(r.Context(), "handlers: rendering collection", "error", err)
		}
	}
}
