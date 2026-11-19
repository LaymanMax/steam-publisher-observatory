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

	filter := analytics.PublisherFilter{
		Query:     q.Get("query"),
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

	filter := analytics.CatalogFilter{
		Query:           q.Get("query"),
		Genre:           q.Get("genre"),
		DeveloperID:     q.Get("developer_id"),
		PriceTier:       q.Get("price_tier"),
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
