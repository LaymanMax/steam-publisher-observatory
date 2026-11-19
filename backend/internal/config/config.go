package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// PriceTier defines a pricing tier threshold.
type PriceTier struct {
	Name string  `json:"name"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// Config holds all backend runtime configuration.
type Config struct {
	Port         int
	Host         string
	DataDir      string
	DBPath       string
	HFRepoID     string
	CORSAllowAll bool
	PriceTiers   []PriceTier
}

// DefaultPriceTiers defines standard pricing brackets in USD.
var DefaultPriceTiers = []PriceTier{
	{Name: "Free", Min: 0.0, Max: 0.0},
	{Name: "Very Low", Min: 0.01, Max: 4.99},
	{Name: "Low", Min: 5.00, Max: 9.99},
	{Name: "Mid-range", Min: 10.00, Max: 19.99},
	{Name: "Premium", Min: 20.00, Max: 49.99},
	{Name: "High-priced", Min: 50.00, Max: 99999.0},
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			port = parsed
		}
	}

	host := "127.0.0.1"
	if h := os.Getenv("HOST"); h != "" {
		host = h
	}

	dataDir := "./data"
	if d := os.Getenv("DATA_DIR"); d != "" {
		dataDir = d
	}
	_ = os.MkdirAll(dataDir, 0755)

	dbPath := filepath.Join(dataDir, "observatory.duckdb")
	if db := os.Getenv("DB_PATH"); db != "" {
		dbPath = db
	}

	hfRepo := "FronkonGames/steam-games-dataset"
	if r := os.Getenv("HF_REPO_ID"); r != "" {
		hfRepo = r
	}

	return &Config{
		Port:         port,
		Host:         host,
		DataDir:      dataDir,
		DBPath:       dbPath,
		HFRepoID:     hfRepo,
		CORSAllowAll: true,
		PriceTiers:   DefaultPriceTiers,
	}
}
