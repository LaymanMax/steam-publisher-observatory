# Dataset Acquisition & Ingestion Guide

## 1. Primary Source Overview

The Steam Publisher Observatory uses the public snapshot dataset from Hugging Face:

- **Repository**: [`FronkonGames/steam-games-dataset`](https://huggingface.co/datasets/FronkonGames/steam-games-dataset)
- **Format**: Apache Parquet (`snappy` compressed)
- **Size**: Approximately 183 MB compressed (~150 MB DuckDB relational footprint)
- **Total Titles**: 124,146 published games

---

## 2. Ingestion Options

### Option A: Direct Download via Hugging Face Hub (UI or API)

1. Navigate to the **Dataset Manager** (`/dataset`) in the web UI.
2. Click **Start Hugging Face Download & Ingestion**.
3. The Go backend:
   - Queries Hugging Face Hub API `https://huggingface.co/api/datasets/FronkonGames/steam-games-dataset` to dynamically discover available Parquet files.
   - Downloads `train-00000-of-00001.parquet` into `./data/`.
   - Computes and verifies the SHA-256 checksum (`775b17f2a5a52b96a8441739cd0dd4939916335b6edeebbfc4a1b238388c5a77`).
   - Executes the high-speed vectorized DuckDB ingestion pipeline.
   - Live progress, bytes downloaded, and status logs update continuously via `/api/dataset/status`.

### Option B: Local Parquet File Upload

If you have an offline copy of the Parquet dataset:

1. In the **Dataset Manager** page, drag and drop the `.parquet` file into the upload zone.
2. Or use `curl` to upload via multipart form:
   ```bash
   curl -F "file=@/path/to/steam_games.parquet" http://127.0.0.1:8080/api/dataset/upload
   ```

### Option C: Standalone CLI Ingestion

For batch automated processing or continuous integration:

```bash
cd backend
go run cmd/ingest/main.go -parquet ../data/train-00000-of-00001.parquet -db ../data/observatory.duckdb
```

Execution log:

```text
2026/11/19 09:30:11 Starting CLI ingestion from ../data/train-00000-of-00001.parquet into ../data/observatory.duckdb...
2026/11/19 09:30:11 Opening DuckDB database at ../data/observatory.duckdb...
2026/11/19 09:30:11 Initializing database schema...
2026/11/19 09:30:11 Executing vectorized Parquet import from ../data/train-00000-of-00001.parquet...
2026/11/19 09:30:17 Ingestion completed successfully in 6.05s!
2026/11/19 09:30:17 Total games imported: 124,146
2026/11/19 09:30:17 Distinct publishers: 63,070
2026/11/19 09:30:17 Distinct developers: 75,731
```

---

## 3. High-Speed Vectorized Ingestion Pipeline

Traditional Go batch inserts using `database/sql` prepared statements suffer from significant parameter binding overhead (over 80,000 parameters per batch).

Steam Publisher Observatory achieves a **6-second ingestion time for 124,146 records** through a two-stage vectorized pipeline:

1. **Parquet Vector Ingestion (`raw_import`)**:
   DuckDB's native C++ Parquet reader scans the file directly into a temporary analytical staging table `raw_import` at NVMe memory speed.
2. **Relational Vector Normalization**:
   High-performance SQL transformations normalize entities directly inside DuckDB:
   ```sql
   INSERT INTO publishers (publisher_id, display_name, canonical_key, catalog_size)
   SELECT
       md5('publisher:' || LOWER(TRIM(pub))),
       MIN(TRIM(pub)),
       LOWER(TRIM(pub)),
       COUNT(DISTINCT appid)
   FROM (
       SELECT appid, unnest(string_split(publishers, ',')) as pub
       FROM raw_import
       WHERE publishers IS NOT NULL AND TRIM(publishers) != ''
   )
   GROUP BY LOWER(TRIM(pub));
   ```
3. **Data Quality & Playtest Classification**:
   Vectorized pattern matching identifies playtest builds and sparse metadata:
   ```sql
   UPDATE games SET
       is_playtest = CASE
           WHEN LOWER(name) LIKE '%playtest%' OR LOWER(name) LIKE '%closed beta%' THEN true
           ELSE false
       END,
       is_sparse = CASE
           WHEN (short_description IS NULL OR short_description = '')
            AND price = 0
            AND owners_upper = 0 THEN true
           ELSE false
       END;
   ```
4. **Summary Materialization**:
   Pre-computes HHI, effective title counts, and price distributions into `publisher_summaries` for instantaneous sub-millisecond retrieval.
