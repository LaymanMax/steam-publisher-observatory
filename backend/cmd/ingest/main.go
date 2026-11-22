package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/database"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
)

func main() {
	cfg := config.Load()
	defaultParquet := filepath.Join(cfg.DataDir, "train-00000-of-00001.parquet")

	parquetFlag := flag.String("parquet", "", "Path to input Parquet dataset file")
	dbFlag := flag.String("db", cfg.DBPath, "Path to DuckDB database file")
	dataDirFlag := flag.String("datadir", cfg.DataDir, "Directory for dataset and database files")
	flag.Parse()

	cfg.DataDir = *dataDirFlag
	cfg.DBPath = *dbFlag

	targetPath := defaultParquet
	if *parquetFlag != "" {
		targetPath = *parquetFlag
	} else if flag.NArg() > 0 {
		targetPath = flag.Arg(0)
	}

	log.Printf("Connecting to DuckDB database at %s...", cfg.DBPath)
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	imp := importer.NewImporter(db, cfg)

	log.Printf("Starting ingestion of %s...", targetPath)
	start := time.Now()

	stats, err := imp.Ingest(context.Background(), targetPath, "local:"+filepath.Base(targetPath), func(stage importer.ImportStage, pct float64, msg string) {
		log.Printf("[%s] %.1f%% - %s", stage, pct, msg)
	})
	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n=========================================\n")
	fmt.Printf("Dataset Ingestion Succeeded in %v\n", elapsed)
	fmt.Printf("=========================================\n")
	fmt.Printf("Total Rows:              %d\n", stats.RowCount)
	fmt.Printf("Unique AppIDs:           %d\n", stats.UniqueAppIDCount)
	fmt.Printf("Duplicate AppIDs:        %d\n", stats.DuplicateCount)
	fmt.Printf("Invalid Records:         %d\n", stats.InvalidRecordCount)
	fmt.Printf("Missing Publishers:      %d\n", stats.MissingPublisherCount)
	fmt.Printf("Missing Developers:      %d\n", stats.MissingDeveloperCount)
	fmt.Printf("Missing Genres:          %d\n", stats.MissingGenreCount)
	fmt.Printf("Missing Release Dates:   %d\n", stats.MissingReleaseDateCount)
	fmt.Printf("Missing Prices:          %d\n", stats.MissingPriceCount)
	fmt.Printf("Missing Owners:          %d\n", stats.MissingOwnersCount)
	fmt.Printf("Suspected Playtests:     %d\n", stats.SuspectedPlaytestCount)
	fmt.Printf("SHA-256 Checksum:        %s\n", stats.SHA256Checksum)
	fmt.Printf("File Size:               %.2f MB\n", float64(stats.FileSizeBytes)/(1024*1024))
	fmt.Printf("=========================================\n")
}
