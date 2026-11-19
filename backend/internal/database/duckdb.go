package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/marcboeker/go-duckdb"
)

// DB wraps the DuckDB connection pool.
type DB struct {
	*sql.DB
	path string
	mu   sync.RWMutex
}

// Open initializes and connects to the DuckDB database at dbPath.
// If dbPath is empty or ":memory:", an in-memory database is used.
func Open(dbPath string) (*DB, error) {
	connStr := dbPath
	if connStr == "" {
		connStr = ":memory:"
	}

	sqlDB, err := sql.Open("duckdb", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb at %s: %w", connStr, err)
	}

	// Set connection limits appropriate for DuckDB embedded engine
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	db := &DB{
		DB:   sqlDB,
		path: dbPath,
	}

	if err := db.InitSchema(context.Background()); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// InitSchema creates the normalized relational and analytical tables if they do not exist.
func (db *DB) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS dataset_imports (
			id VARCHAR PRIMARY KEY,
			source VARCHAR NOT NULL,
			imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			row_count BIGINT DEFAULT 0,
			unique_appid_count BIGINT DEFAULT 0,
			duplicate_count BIGINT DEFAULT 0,
			invalid_record_count BIGINT DEFAULT 0,
			missing_publisher_count BIGINT DEFAULT 0,
			missing_developer_count BIGINT DEFAULT 0,
			missing_genre_count BIGINT DEFAULT 0,
			missing_price_count BIGINT DEFAULT 0,
			missing_release_date_count BIGINT DEFAULT 0,
			missing_owners_count BIGINT DEFAULT 0,
			suspected_playtest_count BIGINT DEFAULT 0,
			file_size_bytes BIGINT DEFAULT 0,
			sha256_checksum VARCHAR,
			status VARCHAR DEFAULT 'completed',
			schema_version VARCHAR DEFAULT '1.0'
		);`,

		`CREATE TABLE IF NOT EXISTS games (
			appid VARCHAR PRIMARY KEY,
			name VARCHAR,
			release_date_raw VARCHAR,
			release_date VARCHAR,
			release_year INTEGER,
			has_valid_date BOOLEAN DEFAULT FALSE,
			price DOUBLE,
			price_tier VARCHAR,
			estimated_owners_raw VARCHAR,
			owners_lower BIGINT,
			owners_upper BIGINT,
			owners_midpoint BIGINT,
			has_valid_owners BOOLEAN DEFAULT FALSE,
			peak_ccu BIGINT DEFAULT 0,
			required_age BIGINT DEFAULT 0,
			dlc_count BIGINT DEFAULT 0,
			short_description VARCHAR,
			header_image VARCHAR,
			website VARCHAR,
			support_url VARCHAR,
			windows BOOLEAN DEFAULT FALSE,
			mac BOOLEAN DEFAULT FALSE,
			linux BOOLEAN DEFAULT FALSE,
			metacritic_score BIGINT,
			user_score BIGINT,
			positive BIGINT DEFAULT 0,
			negative BIGINT DEFAULT 0,
			total_reviews BIGINT DEFAULT 0,
			positive_share DOUBLE,
			reviews_raw VARCHAR,
			achievements BIGINT DEFAULT 0,
			recommendations BIGINT DEFAULT 0,
			average_playtime_forever BIGINT DEFAULT 0,
			average_playtime_2weeks BIGINT DEFAULT 0,
			median_playtime_forever BIGINT DEFAULT 0,
			median_playtime_2weeks BIGINT DEFAULT 0,
			is_playtest BOOLEAN DEFAULT FALSE,
			is_sparse BOOLEAN DEFAULT FALSE,
			quality_score DOUBLE DEFAULT 0.0,
			quality_flags VARCHAR
		);`,

		`CREATE TABLE IF NOT EXISTS publishers (
			id VARCHAR PRIMARY KEY,
			name VARCHAR NOT NULL,
			normalized_key VARCHAR NOT NULL,
			game_count BIGINT DEFAULT 0,
			alias_of_id VARCHAR
		);`,

		`CREATE TABLE IF NOT EXISTS developers (
			id VARCHAR PRIMARY KEY,
			name VARCHAR NOT NULL,
			normalized_key VARCHAR NOT NULL,
			game_count BIGINT DEFAULT 0,
			alias_of_id VARCHAR
		);`,

		`CREATE TABLE IF NOT EXISTS game_publishers (
			appid VARCHAR NOT NULL,
			publisher_id VARCHAR NOT NULL,
			PRIMARY KEY (appid, publisher_id)
		);`,

		`CREATE TABLE IF NOT EXISTS game_developers (
			appid VARCHAR NOT NULL,
			developer_id VARCHAR NOT NULL,
			PRIMARY KEY (appid, developer_id)
		);`,

		`CREATE TABLE IF NOT EXISTS game_genres (
			appid VARCHAR NOT NULL,
			genre VARCHAR NOT NULL,
			PRIMARY KEY (appid, genre)
		);`,

		`CREATE TABLE IF NOT EXISTS game_categories (
			appid VARCHAR NOT NULL,
			category VARCHAR NOT NULL,
			PRIMARY KEY (appid, category)
		);`,

		`CREATE TABLE IF NOT EXISTS entity_aliases (
			id VARCHAR PRIMARY KEY,
			entity_type VARCHAR NOT NULL,
			source_name VARCHAR NOT NULL,
			canonical_id VARCHAR NOT NULL,
			canonical_name VARCHAR NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			method VARCHAR NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS publisher_summaries (
			publisher_id VARCHAR PRIMARY KEY,
			name VARCHAR NOT NULL,
			catalog_size BIGINT DEFAULT 0,
			developer_count BIGINT DEFAULT 0,
			genre_count BIGINT DEFAULT 0,
			earliest_year INTEGER,
			latest_year INTEGER,
			valid_owners_count BIGINT DEFAULT 0,
			ownership_coverage DOUBLE DEFAULT 0.0,
			catalog_owners_lower BIGINT DEFAULT 0,
			catalog_owners_midpoint BIGINT DEFAULT 0,
			catalog_owners_upper BIGINT DEFAULT 0,
			top_title_name VARCHAR,
			top_title_midpoint BIGINT DEFAULT 0,
			top1_share DOUBLE DEFAULT 0.0,
			top3_share DOUBLE DEFAULT 0.0,
			top5_share DOUBLE DEFAULT 0.0,
			top10_share DOUBLE DEFAULT 0.0,
			remaining_share DOUBLE DEFAULT 0.0,
			hhi DOUBLE DEFAULT 0.0,
			effective_title_count DOUBLE DEFAULT 0.0,
			free_title_count BIGINT DEFAULT 0,
			free_title_share DOUBLE DEFAULT 0.0,
			median_price DOUBLE DEFAULT 0.0,
			avg_price DOUBLE DEFAULT 0.0,
			total_positive_reviews BIGINT DEFAULT 0,
			total_negative_reviews BIGINT DEFAULT 0,
			total_reviews BIGINT DEFAULT 0,
			avg_positive_share DOUBLE DEFAULT 0.0,
			median_playtime BIGINT DEFAULT 0,
			avg_peak_ccu DOUBLE DEFAULT 0.0,
			max_peak_ccu BIGINT DEFAULT 0,
			dominant_genre VARCHAR,
			genre_concentration DOUBLE DEFAULT 0.0,
			avg_quality_score DOUBLE DEFAULT 0.0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error executing query: %s: %w", query, err)
		}
	}

	return nil
}

// LockForWrite acquires write lock for modifying entity relationships.
func (db *DB) LockForWrite() {
	db.mu.Lock()
}

// UnlockForWrite releases write lock.
func (db *DB) UnlockForWrite() {
	db.mu.Unlock()
}
