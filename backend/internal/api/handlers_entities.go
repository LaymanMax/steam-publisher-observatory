package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CreateAliasRequest struct {
	EntityType  string `json:"entity_type"` // "publisher" or "developer"
	SourceName  string `json:"source_name"`
	CanonicalID string `json:"canonical_id"`
}

func (s *Server) handleListAliases(w http.ResponseWriter, r *http.Request) {
	aliases, err := s.analytics.ListAliases(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"aliases": aliases,
	})
}

func (s *Server) handleCreateAlias(w http.ResponseWriter, r *http.Request) {
	var req CreateAliasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if req.SourceName == "" || req.CanonicalID == "" {
		respondError(w, http.StatusBadRequest, "source_name and canonical_id are required")
		return
	}
	if req.EntityType == "" {
		req.EntityType = "publisher"
	}

	err := s.analytics.CreateAlias(r.Context(), req.EntityType, req.SourceName, req.CanonicalID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Alias created and metrics recalculated successfully.",
	})
}

func (s *Server) handleDeleteAlias(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Alias ID required")
		return
	}

	if err := s.analytics.RemoveAlias(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Alias removed successfully.",
	})
}

func (s *Server) handleListPotentialAliases(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	pairs, err := s.analytics.ListPotentialAliases(r.Context(), limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"potential_aliases": pairs,
	})
}

func (s *Server) handleListDevelopers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("query")
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	devs, total, err := s.analytics.ListDevelopers(r.Context(), query, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"developers": devs,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}
