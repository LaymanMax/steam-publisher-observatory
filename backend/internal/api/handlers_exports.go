package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/reports"
)

func (s *Server) handleExportPublishersCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	minGames, _ := strconv.Atoi(q.Get("min_games"))

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
		Limit:     1000,
		Offset:    0,
	}

	publishers, _, err := s.analytics.ListPublishers(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	csvData, err := reports.ExportPublishersCSV(publishers)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"publishers_export.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (s *Server) handleExportCatalogCSV(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	publisherID := q.Get("publisher_id")
	if publisherID == "" {
		respondError(w, http.StatusBadRequest, "publisher_id is required")
		return
	}

	excludePlaytest := strings.EqualFold(q.Get("exclude_playtest"), "true") || q.Get("exclude_playtest") == "1"

	query := q.Get("query")
	if query == "" {
		query = q.Get("q")
	}

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
		Limit:           5000,
		Offset:          0,
	}

	games, _, err := s.analytics.GetPublisherCatalog(r.Context(), publisherID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	csvData, err := reports.ExportCatalogCSV(games)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("catalog_%s.csv", publisherID)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (s *Server) handleExportReportJSON(w http.ResponseWriter, r *http.Request) {
	publisherID := r.URL.Query().Get("publisher_id")
	if publisherID == "" {
		respondError(w, http.StatusBadRequest, "publisher_id is required")
		return
	}

	jsonData, err := reports.GeneratePublisherJSONReport(r.Context(), s.analytics, publisherID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("report_%s.json", publisherID)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonData)
}
