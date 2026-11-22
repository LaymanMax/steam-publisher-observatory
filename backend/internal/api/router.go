package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/dataset"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/jobs"
)

// Server holds dependencies for API routing and request handling.
type Server struct {
	config    *config.Config
	analytics *analytics.Service
	importer  *importer.Importer
	jobs      *jobs.Manager
	hfClient  *dataset.Client
}

// NewServer constructs the API Server.
func NewServer(
	cfg *config.Config,
	svc *analytics.Service,
	imp *importer.Importer,
	jobMgr *jobs.Manager,
	hf *dataset.Client,
) *Server {
	return &Server{
		config:    cfg,
		analytics: svc,
		importer:  imp,
		jobs:      jobMgr,
		hfClient:  hf,
	}
}

// Router builds and configures the HTTP router with middlewares and route handlers.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration for local development & production
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"product": "Steam Publisher Observatory",
			"version": "1.0.0",
		})
	})

	// Main overview
	r.Get("/api/overview", s.handleGetOverview)

	// Dataset management
	r.Route("/api/dataset", func(r chi.Router) {
		r.Get("/status", s.handleDatasetStatus)
		r.Get("/imports", s.handleDatasetImports)
		r.Post("/import", s.handleDatasetImport)
		r.Post("/upload", s.handleDatasetUpload)
		r.Post("/cancel", s.handleDatasetCancel)
		r.Get("/schema", s.handleDatasetSchema)
		r.Get("/quality", s.handleDatasetQuality)
		r.Get("/issues", s.handleDatasetIssues)
	})

	// Publishers
	r.Route("/api/publishers", func(r chi.Router) {
		r.Get("/", s.handleListPublishers)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", s.handleGetPublisher)
			r.Get("/summary", s.handleGetPublisherSummary)
			r.Get("/catalog", s.handleGetPublisherCatalog)
			r.Get("/genres", s.handleGetPublisherGenres)
			r.Get("/prices", s.handleGetPublisherPrices)
			r.Get("/timeline", s.handleGetPublisherTimeline)
			r.Get("/developers", s.handleGetPublisherDevelopers)
			r.Get("/dependency", s.handleGetPublisherDependency)
			r.Get("/outliers", s.handleGetPublisherOutliers)
			r.Get("/quality", s.handleGetPublisherQuality)
			r.Get("/report", s.handleGetPublisherReport)
		})
	})

	// Multi-publisher comparison
	r.Get("/api/compare", s.handleCompare)

	// Entities and aliases
	r.Route("/api/entities", func(r chi.Router) {
		r.Get("/publishers", s.handleListPublishers)
		r.Get("/developers", s.handleListDevelopers)
		r.Get("/aliases", s.handleListAliases)
		r.Post("/aliases", s.handleCreateAlias)
		r.Delete("/aliases/{id}", s.handleDeleteAlias)
		r.Get("/potential-aliases", s.handleListPotentialAliases)
	})

	// Exports
	r.Route("/api/exports", func(r chi.Router) {
		r.Get("/publishers.csv", s.handleExportPublishersCSV)
		r.Get("/catalog.csv", s.handleExportCatalogCSV)
		r.Get("/report.json", s.handleExportReportJSON)
	})

	// Static SPA frontend serving if built dist folder exists
	distDirs := []string{"./frontend/dist", "../frontend/dist", "./dist"}
	for _, d := range distDirs {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			distDir := d
			r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
				target := filepath.Join(distDir, filepath.Clean(req.URL.Path))
				info, err := os.Stat(target)
				if os.IsNotExist(err) || (err == nil && info.IsDir()) {
					target = filepath.Join(distDir, "index.html")
					info, err = os.Stat(target)
				}
				if err != nil {
					http.NotFound(w, req)
					return
				}
				data, err := os.ReadFile(target)
				if err != nil {
					http.Error(w, "failed to read static asset", http.StatusInternalServerError)
					return
				}
				http.ServeContent(w, req, info.Name(), info.ModTime(), bytes.NewReader(data))
			})
			break
		}
	}

	return r
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
		"success": false,
	})
}
