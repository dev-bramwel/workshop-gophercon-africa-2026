package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"uuid"
	"workshop/internal/domain/sighting"
	"workshop/web/views/pages"
)

type SightingsFinder interface {
	Find(context.Context, uuid.UUID) (sighting.Sighting, error)
}

func Sighting(logger *slog.Logger, svc SightingsFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			renderError(r.Context(), w, logger, 404, "We couldn’t find this sighting.")
			return
		}
		s, err := svc.Find(r.Context(), id)
		if errors.Is(err, sighting.ErrNotFound) {
			renderError(r.Context(), w, logger, 404, "We couldn’t find this sighting.")
			return
		}
		if err != nil {
			logger.ErrorContext(r.Context(), "handlers: finding sighting", "error", err)
			renderError(r.Context(), w, logger, 500, "Unable to load this sighting.")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pages.SightingDetail(s).Render(r.Context(), w); err != nil {
			logger.ErrorContext(r.Context(), "handlers: rendering sighting", "error", err)
		}
	}
}
func Animal(logger *slog.Logger, svc SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || key <= 0 {
			renderError(r.Context(), w, logger, 404, "We couldn’t find this animal.")
			return
		}
		items, err := svc.List(r.Context(), sighting.ListFilter{})
		if err != nil {
			logger.ErrorContext(r.Context(), "handlers: finding animal", "error", err)
			renderError(r.Context(), w, logger, 500, "Unable to load this animal.")
			return
		}
		d := pages.NewCollection(items, "/gallery", strconv.Itoa(key), "", "", "", 1)
		if len(d.Species) == 0 {
			renderError(r.Context(), w, logger, 404, "We couldn’t find this animal.")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := pages.AnimalDetail(d.Species[0], d).Render(r.Context(), w); err != nil {
			logger.ErrorContext(r.Context(), "handlers: rendering animal", "error", err)
		}
	}
}
