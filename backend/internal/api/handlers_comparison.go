package api

import (
	"net/http"
	"strings"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/comparison"
)

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rawIDs := q.Get("publishers")
	if rawIDs == "" {
		rawIDs = q.Get("ids")
	}

	var publisherIDs []string
	if rawIDs != "" {
		parts := strings.Split(rawIDs, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				publisherIDs = append(publisherIDs, trimmed)
			}
		}
	} else {
		// Alternatively check array param like publishers=A&publishers=B
		publisherIDs = q["publishers"]
	}

	if len(publisherIDs) == 0 {
		respondError(w, http.StatusBadRequest, "At least one publisher ID must be specified via ?publishers=id1,id2")
		return
	}

	result, err := comparison.Compare(r.Context(), s.analytics, publisherIDs)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}
