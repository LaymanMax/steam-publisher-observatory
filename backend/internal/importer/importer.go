package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/database"
)

// ImportStage represents the current phase of the ingestion job.
type ImportStage string

const (
	StageValidating  ImportStage = "validating"
	StageIngesting   ImportStage = "ingesting"
	StageNormalizing ImportStage = "normalizing"
	StageSummarizing ImportStage = "summarizing"
	StageComplete    ImportStage = "complete"
	StageFailed      ImportStage = "failed"
)

// ProgressCallback reports ingestion progress.
type StageProgressCallback func(stage ImportStage, percent float64, message string)

// ImportStats captures the complete diagnostics of an imported dataset.
type ImportStats struct {
	ImportID                string    `json:"import_id"`
	Source                  string    `json:"source"`
	ImportedAt              time.Time `json:"imported_at"`
	RowCount                int64     `json:"row_count"`
	UniqueAppIDCount        int64     `json:"unique_appid_count"`
	DuplicateCount          int64     `json:"duplicate_count"`
	InvalidRecordCount      int64     `json:"invalid_record_count"`
	MissingPublisherCount   int64     `json:"missing_publisher_count"`
	MissingDeveloperCount   int64     `json:"missing_developer_count"`
	MissingGenreCount       int64     `json:"missing_genre_count"`
	MissingPriceCount       int64     `json:"missing_price_count"`
	MissingReleaseDateCount int64     `json:"missing_release_date_count"`
	MissingOwnersCount      int64     `json:"missing_owners_count"`
	SuspectedPlaytestCount  int64     `json:"suspected_playtest_count"`
	FileSizeBytes           int64     `json:"file_size_bytes"`
	SHA256Checksum          string    `json:"sha256_checksum"`
	Status                  string    `json:"status"`
}

// Importer orchestrates Parquet file inspection and ingestion.
type Importer struct {
	db     *database.DB
	config *config.Config
}

// NewImporter constructs an Importer instance.
func NewImporter(db *database.DB, cfg *config.Config) *Importer {
	return &Importer{
		db:     db,
		config: cfg,
	}
}

// InspectParquet validates that the file is readable and inspects its schema and row count.
func (imp *Importer) InspectParquet(ctx context.Context, filePath string) (int64, []string, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM read_parquet('%s')", filePath)
	if err := imp.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, nil, fmt.Errorf("failed to read parquet file: %w", err)
	}

	describeQuery := fmt.Sprintf("DESCRIBE SELECT * FROM read_parquet('%s')", filePath)
	rows, err := imp.db.QueryContext(ctx, describeQuery)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to inspect parquet schema: %w", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var colName, colType, isNull, key, dflt, extra sql.NullString
		if err := rows.Scan(&colName, &colType, &isNull, &key, &dflt, &extra); err == nil {
			columns = append(columns, colName.String)
		}
	}

	return count, columns, nil
}

// Ingest imports the Parquet file into the database using DuckDB vectorized columnar loading.
func (imp *Importer) Ingest(ctx context.Context, filePath, sourceName string, onProgress StageProgressCallback) (*ImportStats, error) {
	imp.db.LockForWrite()
	defer imp.db.UnlockForWrite()

	importID := uuid.New().String()
	startTime := time.Now()

	// 1. Calculate file size and checksum
	if onProgress != nil {
		onProgress(StageValidating, 5.0, "Calculating file checksum and verifying Parquet structure...")
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file error: %w", err)
	}
	fileSize := fi.Size()

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for checksum: %w", err)
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		f.Close()
		return nil, fmt.Errorf("checksum calculation error: %w", err)
	}
	f.Close()
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// 2. Validate Parquet & inspect row count
	rowCount, columns, err := imp.InspectParquet(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("parquet validation failed: %w", err)
	}
	if rowCount == 0 {
		return nil, fmt.Errorf("parquet file contains 0 rows")
	}

	if onProgress != nil {
		onProgress(StageValidating, 15.0, fmt.Sprintf("Parquet schema verified (%d columns, %d rows)", len(columns), rowCount))
	}

	// 3. Staging table
	if onProgress != nil {
		onProgress(StageIngesting, 25.0, "Loading records into analytical staging...")
	}

	cleanQueries := []string{
		"DELETE FROM game_publishers",
		"DELETE FROM game_developers",
		"DELETE FROM game_genres",
		"DELETE FROM game_categories",
		"DELETE FROM games",
		"DELETE FROM publishers",
		"DELETE FROM developers",
		"DELETE FROM publisher_summaries",
	}
	for _, cq := range cleanQueries {
		if _, err := imp.db.ExecContext(ctx, cq); err != nil {
			return nil, fmt.Errorf("failed to clean working tables: %w", err)
		}
	}

	createTemp := fmt.Sprintf("CREATE OR REPLACE TEMP TABLE raw_import AS SELECT * FROM read_parquet('%s')", filePath)
	if _, err := imp.db.ExecContext(ctx, createTemp); err != nil {
		return nil, fmt.Errorf("failed to create staging table: %w", err)
	}

	// 4. Ingest normalized Publishers
	if onProgress != nil {
		onProgress(StageNormalizing, 40.0, "Normalizing publisher entities...")
	}
	insertPublishersSQL := `
		INSERT OR IGNORE INTO publishers (id, name, normalized_key, game_count)
		SELECT 
			md5('publisher:' || LOWER(TRIM(pub))) as id,
			MIN(TRIM(pub)) as name,
			LOWER(TRIM(pub)) as normalized_key,
			COUNT(DISTINCT appID) as game_count
		FROM (
			SELECT appID, UNNEST(publishers) as pub 
			FROM raw_import 
			WHERE publishers IS NOT NULL
		)
		WHERE TRIM(pub) != ''
		GROUP BY LOWER(TRIM(pub))
	`
	if _, err := imp.db.ExecContext(ctx, insertPublishersSQL); err != nil {
		return nil, fmt.Errorf("failed to ingest publishers: %w", err)
	}

	// 5. Ingest normalized Developers
	if onProgress != nil {
		onProgress(StageNormalizing, 50.0, "Normalizing developer entities...")
	}
	insertDevelopersSQL := `
		INSERT OR IGNORE INTO developers (id, name, normalized_key, game_count)
		SELECT 
			md5('developer:' || LOWER(TRIM(dev))) as id,
			MIN(TRIM(dev)) as name,
			LOWER(TRIM(dev)) as normalized_key,
			COUNT(DISTINCT appID) as game_count
		FROM (
			SELECT appID, UNNEST(developers) as dev 
			FROM raw_import 
			WHERE developers IS NOT NULL
		)
		WHERE TRIM(dev) != ''
		GROUP BY LOWER(TRIM(dev))
	`
	if _, err := imp.db.ExecContext(ctx, insertDevelopersSQL); err != nil {
		return nil, fmt.Errorf("failed to ingest developers: %w", err)
	}

	// 6. Ingest Relationships (game_publishers, game_developers, game_genres, game_categories)
	if onProgress != nil {
		onProgress(StageNormalizing, 60.0, "Indexing relationships across publishers, developers, and genres...")
	}
	relQueries := []string{
		`INSERT OR IGNORE INTO game_publishers (appid, publisher_id)
		SELECT DISTINCT
			CAST(appID AS VARCHAR),
			md5('publisher:' || LOWER(TRIM(pub))) as publisher_id
		FROM (
			SELECT appID, UNNEST(publishers) as pub 
			FROM raw_import 
			WHERE publishers IS NOT NULL
		)
		WHERE TRIM(pub) != '' AND appID IS NOT NULL`,

		`INSERT OR IGNORE INTO game_developers (appid, developer_id)
		SELECT DISTINCT
			CAST(appID AS VARCHAR),
			md5('developer:' || LOWER(TRIM(dev))) as developer_id
		FROM (
			SELECT appID, UNNEST(developers) as dev 
			FROM raw_import 
			WHERE developers IS NOT NULL
		)
		WHERE TRIM(dev) != '' AND appID IS NOT NULL`,

		`INSERT OR IGNORE INTO game_genres (appid, genre)
		SELECT DISTINCT
			CAST(appID AS VARCHAR),
			TRIM(g) as genre
		FROM (
			SELECT appID, UNNEST(genres) as g 
			FROM raw_import 
			WHERE genres IS NOT NULL
		)
		WHERE TRIM(g) != '' AND appID IS NOT NULL`,

		`INSERT OR IGNORE INTO game_categories (appid, category)
		SELECT DISTINCT
			CAST(appID AS VARCHAR),
			TRIM(c) as category
		FROM (
			SELECT appID, UNNEST(categories) as c 
			FROM raw_import 
			WHERE categories IS NOT NULL
		)
		WHERE TRIM(c) != '' AND appID IS NOT NULL`,
	}
	for _, rq := range relQueries {
		if _, err := imp.db.ExecContext(ctx, rq); err != nil {
			return nil, fmt.Errorf("failed to index relationships: %w", err)
		}
	}

	// 7. Ingest Games with Data Quality & Playtest Heuristics
	if onProgress != nil {
		onProgress(StageIngesting, 70.0, "Transforming and persisting normalized game records...")
	}
	insertGamesSQL := `
		INSERT OR REPLACE INTO games
		WITH parsed AS (
			SELECT 
				CAST(appID AS VARCHAR) as appid,
				COALESCE(name, '') as name,
				release_date as release_date_raw,
				TRY_CAST(REGEXP_EXTRACT(release_date, '\b(19[89]\d|20[0-3]\d)\b') AS INTEGER) as release_year,
				CASE 
					WHEN TRY_CAST(REGEXP_EXTRACT(release_date, '\b(19[89]\d|20[0-3]\d)\b') AS INTEGER) IS NOT NULL THEN TRUE 
					ELSE FALSE 
				END as has_valid_date,
				price,
				CASE 
					WHEN price IS NULL OR price < 0 THEN 'Missing'
					WHEN price = 0.0 THEN 'Free'
					WHEN price <= 4.99 THEN 'Very Low'
					WHEN price <= 9.99 THEN 'Low'
					WHEN price <= 19.99 THEN 'Mid-range'
					WHEN price <= 49.99 THEN 'Premium'
					ELSE 'High-priced'
				END as price_tier,
				estimated_owners as estimated_owners_raw,
				CASE 
					WHEN estimated_owners IS NOT NULL AND (estimated_owners LIKE '%-%' OR estimated_owners LIKE '%–%' OR estimated_owners LIKE '%—%' OR estimated_owners LIKE '%−%')
					THEN TRY_CAST(REGEXP_REPLACE(SPLIT_PART(REPLACE(REPLACE(REPLACE(estimated_owners, '–', '-'), '—', '-'), '−', '-'), '-', 1), '[^0-9]', '', 'g') AS BIGINT)
					ELSE NULL 
				END as raw_lower,
				CASE 
					WHEN estimated_owners IS NOT NULL AND (estimated_owners LIKE '%-%' OR estimated_owners LIKE '%–%' OR estimated_owners LIKE '%—%' OR estimated_owners LIKE '%−%')
					THEN TRY_CAST(REGEXP_REPLACE(SPLIT_PART(REPLACE(REPLACE(REPLACE(estimated_owners, '–', '-'), '—', '-'), '−', '-'), '-', 2), '[^0-9]', '', 'g') AS BIGINT)
					ELSE NULL 
				END as raw_upper,
				COALESCE(peak_ccu, 0) as peak_ccu,
				COALESCE(required_age, 0) as required_age,
				COALESCE(dlc_count, 0) as dlc_count,
				COALESCE(short_description, '') as short_description,
				COALESCE(header_image, '') as header_image,
				COALESCE(website, '') as website,
				COALESCE(support_url, '') as support_url,
				COALESCE(windows, FALSE) as windows,
				COALESCE(mac, FALSE) as mac,
				COALESCE(linux, FALSE) as linux,
				metacritic_score,
				user_score,
				COALESCE(positive, 0) as positive,
				COALESCE(negative, 0) as negative,
				COALESCE(positive, 0) + COALESCE(negative, 0) as total_reviews,
				CASE 
					WHEN (COALESCE(positive, 0) + COALESCE(negative, 0)) > 0 
					THEN CAST(COALESCE(positive, 0) AS DOUBLE) / CAST(COALESCE(positive, 0) + COALESCE(negative, 0) AS DOUBLE) * 100.0
					ELSE NULL 
				END as positive_share,
				COALESCE(reviews, '') as reviews_raw,
				COALESCE(achievements, 0) as achievements,
				COALESCE(recommendations, 0) as recommendations,
				COALESCE(average_playtime_forever, 0) as average_playtime_forever,
				COALESCE(average_playtime_2weeks, 0) as average_playtime_2weeks,
				COALESCE(median_playtime_forever, 0) as median_playtime_forever,
				COALESCE(median_playtime_2weeks, 0) as median_playtime_2weeks,
				CASE 
					WHEN LOWER(COALESCE(name, '')) LIKE '%playtest%' OR LOWER(COALESCE(name, '')) LIKE '%beta test%' THEN TRUE 
					ELSE FALSE 
				END as is_playtest,
				publishers,
				developers,
				genres
			FROM raw_import
			WHERE appID IS NOT NULL AND TRIM(CAST(appID AS VARCHAR)) != ''
		),
		bounded AS (
			SELECT 
				*,
				CASE 
					WHEN raw_lower IS NOT NULL AND raw_upper IS NOT NULL THEN LEAST(raw_lower, raw_upper)
					ELSE NULL 
				END as owners_lower,
				CASE 
					WHEN raw_lower IS NOT NULL AND raw_upper IS NOT NULL THEN GREATEST(raw_lower, raw_upper)
					ELSE NULL 
				END as owners_upper,
				CASE 
					WHEN raw_lower IS NOT NULL AND raw_upper IS NOT NULL THEN (raw_lower + raw_upper) // 2
					ELSE NULL 
				END as owners_midpoint,
				CASE 
					WHEN raw_lower IS NOT NULL AND raw_upper IS NOT NULL THEN TRUE 
					ELSE FALSE 
				END as has_valid_owners,
				CASE 
					WHEN (publishers IS NULL OR len(publishers) = 0) 
					 AND (developers IS NULL OR len(developers) = 0) 
					 AND total_reviews = 0 THEN TRUE 
					ELSE FALSE 
				END as is_sparse
			FROM parsed
		)
		SELECT 
			appid, name, release_date_raw, 
			CASE WHEN has_valid_date THEN CAST(release_year AS VARCHAR) ELSE NULL END as release_date,
			release_year, has_valid_date, price, price_tier,
			estimated_owners_raw, owners_lower, owners_upper, owners_midpoint, has_valid_owners,
			peak_ccu, required_age, dlc_count, short_description, header_image, website, support_url,
			windows, mac, linux, metacritic_score, user_score, positive, negative, total_reviews, positive_share,
			reviews_raw, achievements, recommendations, average_playtime_forever, average_playtime_2weeks,
			median_playtime_forever, median_playtime_2weeks, is_playtest, is_sparse,
			(
				CASE WHEN name != '' THEN 10.0 ELSE 0.0 END +
				CASE WHEN publishers IS NOT NULL AND len(publishers) > 0 THEN 20.0 ELSE 0.0 END +
				CASE WHEN developers IS NOT NULL AND len(developers) > 0 THEN 15.0 ELSE 0.0 END +
				CASE WHEN genres IS NOT NULL AND len(genres) > 0 THEN 15.0 ELSE 0.0 END +
				CASE WHEN has_valid_owners THEN 15.0 ELSE 0.0 END +
				CASE WHEN has_valid_date THEN 10.0 ELSE 0.0 END +
				CASE WHEN price IS NOT NULL THEN 5.0 ELSE 0.0 END +
				CASE WHEN total_reviews > 0 THEN 10.0 ELSE 0.0 END
			) as quality_score,
			CONCAT_WS(',',
				CASE WHEN is_playtest THEN 'suspected_playtest' END,
				CASE WHEN is_sparse THEN 'sparse_metadata' END,
				CASE WHEN publishers IS NULL OR len(publishers) = 0 THEN 'missing_publisher' END,
				CASE WHEN developers IS NULL OR len(developers) = 0 THEN 'missing_developer' END,
				CASE WHEN genres IS NULL OR len(genres) = 0 THEN 'missing_genre' END,
				CASE WHEN NOT has_valid_owners THEN 'missing_owners' END,
				CASE WHEN NOT has_valid_date THEN 'missing_release_date' END,
				CASE WHEN price IS NULL THEN 'missing_price' END
			) as quality_flags
		FROM bounded
	`
	if _, err := imp.db.ExecContext(ctx, insertGamesSQL); err != nil {
		return nil, fmt.Errorf("failed to persist normalized games: %w", err)
	}

	// 8. Materialize publisher analytical summaries
	if onProgress != nil {
		onProgress(StageSummarizing, 85.0, "Materializing portfolio dependency indexes & analytics...")
	}
	if err := imp.MaterializePublisherSummaries(ctx); err != nil {
		return nil, fmt.Errorf("failed to materialize publisher summaries: %w", err)
	}

	// 9. Collect quality diagnostics
	stats := &ImportStats{
		ImportID:       importID,
		Source:         sourceName,
		ImportedAt:     startTime,
		RowCount:       rowCount,
		FileSizeBytes:  fileSize,
		SHA256Checksum: checksum,
		Status:         "completed",
	}

	diagSQL := `
		SELECT 
			COUNT(DISTINCT appid),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_publishers) THEN 1 END),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_developers) THEN 1 END),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_genres) THEN 1 END),
			COUNT(CASE WHEN price IS NULL THEN 1 END),
			COUNT(CASE WHEN has_valid_date = FALSE THEN 1 END),
			COUNT(CASE WHEN has_valid_owners = FALSE THEN 1 END),
			COUNT(CASE WHEN is_playtest = TRUE THEN 1 END)
		FROM games
	`
	_ = imp.db.QueryRowContext(ctx, diagSQL).Scan(
		&stats.UniqueAppIDCount,
		&stats.MissingPublisherCount,
		&stats.MissingDeveloperCount,
		&stats.MissingGenreCount,
		&stats.MissingPriceCount,
		&stats.MissingReleaseDateCount,
		&stats.MissingOwnersCount,
		&stats.SuspectedPlaytestCount,
	)
	stats.DuplicateCount = stats.RowCount - stats.UniqueAppIDCount

	// Record completed import metadata
	insertImportQuery := `
		INSERT INTO dataset_imports (
			id, source, imported_at, row_count, unique_appid_count, duplicate_count,
			invalid_record_count, missing_publisher_count, missing_developer_count,
			missing_genre_count, missing_price_count, missing_release_date_count,
			missing_owners_count, suspected_playtest_count, file_size_bytes,
			sha256_checksum, status, schema_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = imp.db.ExecContext(ctx, insertImportQuery,
		stats.ImportID, stats.Source, stats.ImportedAt, stats.RowCount,
		stats.UniqueAppIDCount, stats.DuplicateCount, stats.InvalidRecordCount,
		stats.MissingPublisherCount, stats.MissingDeveloperCount, stats.MissingGenreCount,
		stats.MissingPriceCount, stats.MissingReleaseDateCount, stats.MissingOwnersCount,
		stats.SuspectedPlaytestCount, stats.FileSizeBytes, stats.SHA256Checksum,
		stats.Status, "1.0",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save dataset import metadata: %w", err)
	}

	if onProgress != nil {
		onProgress(StageComplete, 100.0, fmt.Sprintf("Import complete! %d unique games, ready for observatory.", stats.UniqueAppIDCount))
	}

	return stats, nil
}

// MaterializePublisherSummaries calculates analytical portfolio statistics for each publisher.
func (imp *Importer) MaterializePublisherSummaries(ctx context.Context) error {
	query := `
		INSERT OR REPLACE INTO publisher_summaries
		WITH pub_games AS (
			SELECT 
				gp.publisher_id,
				p.name AS publisher_name,
				g.appid,
				g.name AS game_name,
				g.release_year,
				g.price,
				g.has_valid_owners,
				g.owners_lower,
				g.owners_upper,
				g.owners_midpoint,
				g.positive,
				g.negative,
				g.total_reviews,
				g.peak_ccu,
				g.median_playtime_forever,
				g.quality_score
			FROM game_publishers gp
			JOIN publishers p ON gp.publisher_id = p.id
			JOIN games g ON gp.appid = g.appid
		),
		pub_ranked AS (
			SELECT 
				*,
				ROW_NUMBER() OVER(PARTITION BY publisher_id ORDER BY COALESCE(owners_midpoint, 0) DESC, total_reviews DESC) AS rank
			FROM pub_games
		),
		pub_devs AS (
			SELECT 
				gp.publisher_id,
				COUNT(DISTINCT gd.developer_id) AS dev_count
			FROM game_publishers gp
			JOIN game_developers gd ON gp.appid = gd.appid
			GROUP BY gp.publisher_id
		),
		pub_genres AS (
			SELECT 
				gp.publisher_id,
				COUNT(DISTINCT gg.genre) AS genre_count
			FROM game_publishers gp
			JOIN game_genres gg ON gp.appid = gg.appid
			GROUP BY gp.publisher_id
		),
		pub_top_shares AS (
			SELECT 
				publisher_id,
				MAX(CASE WHEN rank = 1 THEN game_name END) AS top_title_name,
				MAX(CASE WHEN rank = 1 THEN owners_midpoint END) AS top_title_midpoint,
				SUM(CASE WHEN rank <= 1 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top1_midpoint,
				SUM(CASE WHEN rank <= 3 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top3_midpoint,
				SUM(CASE WHEN rank <= 5 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top5_midpoint,
				SUM(CASE WHEN rank <= 10 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top10_midpoint
			FROM pub_ranked
			GROUP BY publisher_id
		),
		pub_hhi_prep AS (
			SELECT 
				pr.publisher_id,
				SUM(
					CASE 
						WHEN tot.total_midpoint > 0 AND pr.owners_midpoint IS NOT NULL 
						THEN POWER(CAST(pr.owners_midpoint AS DOUBLE) / CAST(tot.total_midpoint AS DOUBLE), 2)
						ELSE 0.0 
					END
				) AS unscaled_hhi
			FROM pub_ranked pr
			JOIN (
				SELECT publisher_id, SUM(COALESCE(owners_midpoint, 0)) AS total_midpoint 
				FROM pub_games 
				GROUP BY publisher_id
			) tot ON pr.publisher_id = tot.publisher_id
			GROUP BY pr.publisher_id
		)
		SELECT 
			pg.publisher_id,
			MAX(pg.publisher_name),
			COUNT(DISTINCT pg.appid) AS catalog_size,
			COALESCE(MAX(pd.dev_count), 0) AS developer_count,
			COALESCE(MAX(pgn.genre_count), 0) AS genre_count,
			MIN(pg.release_year) AS earliest_year,
			MAX(pg.release_year) AS latest_year,
			COUNT(CASE WHEN pg.has_valid_owners THEN 1 END) AS valid_owners_count,
			CAST(COUNT(CASE WHEN pg.has_valid_owners THEN 1 END) AS DOUBLE) / CAST(COUNT(DISTINCT pg.appid) AS DOUBLE) * 100.0 AS ownership_coverage,
			COALESCE(SUM(pg.owners_lower), 0) AS catalog_owners_lower,
			COALESCE(SUM(pg.owners_midpoint), 0) AS catalog_owners_midpoint,
			COALESCE(SUM(pg.owners_upper), 0) AS catalog_owners_upper,
			MAX(pts.top_title_name) AS top_title_name,
			COALESCE(MAX(pts.top_title_midpoint), 0) AS top_title_midpoint,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top1_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top1_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top3_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top3_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top5_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top5_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top10_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top10_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN (1.0 - (CAST(MAX(pts.top5_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE))) * 100.0 ELSE 0.0 END AS remaining_share,
			COALESCE(MAX(ph.unscaled_hhi), 0.0) * 10000.0 AS hhi,
			CASE WHEN COALESCE(MAX(ph.unscaled_hhi), 0.0) > 0.00001 THEN 1.0 / MAX(ph.unscaled_hhi) ELSE 0.0 END AS effective_title_count,
			COUNT(CASE WHEN pg.price = 0.0 THEN 1 END) AS free_title_count,
			CAST(COUNT(CASE WHEN pg.price = 0.0 THEN 1 END) AS DOUBLE) / CAST(COUNT(DISTINCT pg.appid) AS DOUBLE) * 100.0 AS free_title_share,
			COALESCE(MEDIAN(pg.price), 0.0) AS median_price,
			COALESCE(AVG(pg.price), 0.0) AS avg_price,
			COALESCE(SUM(pg.positive), 0) AS total_positive_reviews,
			COALESCE(SUM(pg.negative), 0) AS total_negative_reviews,
			COALESCE(SUM(pg.total_reviews), 0) AS total_reviews,
			CASE WHEN SUM(pg.total_reviews) > 0 THEN CAST(SUM(pg.positive) AS DOUBLE) / CAST(SUM(pg.total_reviews) AS DOUBLE) * 100.0 ELSE 0.0 END AS avg_positive_share,
			COALESCE(MEDIAN(pg.median_playtime_forever), 0) AS median_playtime,
			COALESCE(AVG(pg.peak_ccu), 0.0) AS avg_peak_ccu,
			COALESCE(MAX(pg.peak_ccu), 0) AS max_peak_ccu,
			'' AS dominant_genre,
			0.0 AS genre_concentration,
			COALESCE(AVG(pg.quality_score), 0.0) AS avg_quality_score,
			CURRENT_TIMESTAMP AS updated_at
		FROM pub_games pg
		LEFT JOIN pub_devs pd ON pg.publisher_id = pd.publisher_id
		LEFT JOIN pub_genres pgn ON pg.publisher_id = pgn.publisher_id
		LEFT JOIN pub_top_shares pts ON pg.publisher_id = pts.publisher_id
		LEFT JOIN pub_hhi_prep ph ON pg.publisher_id = ph.publisher_id
		GROUP BY pg.publisher_id
	`
	_, err := imp.db.ExecContext(ctx, query)
	return err
}
