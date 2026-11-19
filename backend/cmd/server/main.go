package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/api"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/database"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/dataset"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/jobs"
)

func main() {
	log.Println("Starting Steam Publisher Observatory backend...")

	cfg := config.Load()
	log.Printf("Loaded config: Port=%d, Host=%s, DataDir=%s, DBPath=%s",
		cfg.Port, cfg.Host, cfg.DataDir, cfg.DBPath)

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Embedded DuckDB analytical storage initialized.")

	hfClient := dataset.NewClient(cfg.HFRepoID)
	imp := importer.NewImporter(db, cfg)
	jobMgr := jobs.NewManager(imp)
	analyticsSvc := analytics.NewService(db, cfg)

	server := api.NewServer(cfg, analyticsSvc, imp, jobMgr, hfClient)
	handler := server.Router()

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Server shutdown listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Steam Publisher Observatory HTTP API listening on http://%s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failure: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
