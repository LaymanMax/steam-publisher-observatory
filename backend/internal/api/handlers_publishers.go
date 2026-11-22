package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/reports"
)

func (s *Server) handleListPublishers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	minGames, _ := strconv.Atoi(q.Get("min_games"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	query := q.Get("query")
	if query == "" {
		query = q.Get("q")
	}

	filter := analytics.PublisherFilter{
		Query:     query,
		MinGames:  minGames,
		Genre:     q.Get("genre"),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
		Limit:     limit,
		Offset:    offset,
	}

	publishers, total, err := s.analytics.ListPublishers(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"publishers": publishers,
		"total":      total,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	})
}

func (s *Server) handleGetPublisher(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	summary, err := s.analytics.GetPublisherSummary(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (s *Server) handleGetPublisherSummary(w http.ResponseWriter, r *http.Request) {
	s.handleGetPublisher(w, r)
}

func (s *Server) handleGetPublisherCatalog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var minYear, maxYear *int
	if y, err := strconv.Atoi(q.Get("min_year")); err == nil {
		minYear = &y
	}
	if y, err := strconv.Atoi(q.Get("max_year")); err == nil {
		maxYear = &y
	}

	var minReviews *int64
	if rev, err := strconv.ParseInt(q.Get("min_reviews"), 10, 64); err == nil {
		minReviews = &rev
	}

	excludePlaytest := strings.EqualFold(q.Get("exclude_playtest"), "true") || q.Get("exclude_playtest") == "1"

	query := q.Get("query")
	if query == "" {
		query = q.Get("q")
	}

	filter := analytics.CatalogFilter{
		Query:           query,
		Genre:           q.Get("genre"),
		DeveloperID:     q.Get("developer_id"),
		PriceTier:       q.Get("price_tier"),
		MinYear:         minYear,
		MaxYear:         maxYear,
		MinReviews:      minReviews,
		ExcludePlaytest: excludePlaytest,
		SortBy:          q.Get("sort_by"),
		SortOrder:       q.Get("sort_order"),
		Limit:           limit,
		Offset:          offset,
	}

	games, total, err := s.analytics.GetPublisherCatalog(r.Context(), id, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"games":  games,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (s *Server) handleGetPublisherGenres(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherGenres(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherPrices(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherPrices(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherTimeline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherTimeline(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherDevelopers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherDevelopers(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherDependency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherDependency(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherOutliers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	metric := r.URL.Query().Get("metric")
	report, err := s.analytics.GetPublisherOutliers(r.Context(), id, metric)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherQuality(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	report, err := s.analytics.GetPublisherQuality(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (s *Server) handleGetPublisherReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	data, err := reports.GeneratePublisherJSONReport(r.Context(), s.analytics, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
