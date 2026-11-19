# Architecture Overview: Steam Publisher Observatory

## 1. System Philosophy & Objectives

Steam Publisher Observatory is an analytical platform designed to evaluate video game publishers on Steam at the portfolio and corporate level. Rather than analyzing single titles in isolation, the system models catalog concentration, studio collaboration networks, multi-genre specialization, pricing structures, and historical release cadences across 120,000+ commercial titles.

### Core Architectural Principles

1. **Explainable Intelligence over Black-Box Heuristics**: Every metric (such as Portfolio Dependency and Herfindahl-Hirschman Index) provides clear formulas, input ranges, and data-quality caveats.
2. **Columnar Vectorized Analytical Processing**: The analytical engine uses embedded DuckDB to execute aggregation queries and percentiles directly on columnar storage without loading 124,000+ records into Go garbage-collected memory.
3. **Data Quality Transparency**: Sparse records, suspected playtests, missing publisher credits, and unparseable fields are explicitly preserved, classified, and scored rather than discarded.
4. **Offline First & Zero Mandatory Cloud Dependencies**: The system runs entirely locally with an embedded DuckDB database file, requiring no paid APIs, external database clusters, or third-party accounts.

---

## 2. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  React + TypeScript SPA                     │
│    (Tailwind CSS v4, Lucide Icons, Custom SVG Charts)       │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTP / JSON REST APIs
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go Chi API Server                        │
│  - Middleware: CORS, RequestID, RealIP, Recoverer, Logger   │
│  - Handlers: Overview, Publishers, Compare, Quality, Export │
│  - Background Jobs: HuggingFace Downloader, Parquet Importer│
└──────────────┬───────────────────────────────┬──────────────┘
               │                               │
               ▼                               ▼
┌──────────────────────────────┐ ┌────────────────────────────┐
│      Analytics Engine        │ │    Normalization Engine    │
│  - Portfolio Dependency      │ │  - Unicode NFKC & Trimming │
│  - Weighted Genre Allocation │ │  - Canonical Keying        │
│  - Tukey IQR Outlier Fences  │ │  - Deterministic MD5 IDs   │
│  - Multi-Publisher Compare   │ │  - Alias Resolution DB    │
└──────────────┬───────────────┘ └─────────────┬──────────────┘
               │                               │
               └───────────────┬───────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────┐
│              Embedded DuckDB Vectorized Storage             │
│  - Persistent storage: `data/observatory.duckdb`            │
│  - Ingests raw Parquet via native DuckDB C++ parser         │
│  - Columnar indices on AppID, Publisher ID, Developer ID    │
│  - Materialized analytical rollups for sub-millisecond query│
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Component Breakdown

### 3.1 Backend (`backend/`)

- **`cmd/server/main.go`**: Application entrypoint. Parses CLI flags (`-port`, `-host`, `-db`, `-datadir`), initializes DuckDB, registers HTTP routes, and handles graceful POSIX signals (`SIGINT`, `SIGTERM`).
- **`cmd/ingest/main.go`**: Standalone CLI ingestion tool for automated data pipelines and headless CI ingestion.
- **`internal/config/config.go`**: Environment and flag configuration loader.
- **`internal/database/duckdb.go`**: Embedded DuckDB schema manager, table creations, indices, and thread-safe connection pooling via `github.com/marcboeker/go-duckdb`.
- **`internal/importer/`**:
  - `parser.go`: Robust parsing of estimated ownership ranges (supporting en-dashes, em-dashes, reversed intervals, and commas), multi-format dates, price tier classifications, and data quality scoring.
  - `importer.go`: High-speed vectorized ingestion pipeline reading raw Parquet and executing relational normalization into games, publishers, developers, genres, and link tables in ~6 seconds.
- **`internal/dataset/huggingface.go`**: Hugging Face Hub metadata discovery client, SHA-256 integrity verifier, and chunked download manager with progress reporting.
- **`internal/normalization/normalizer.go`**: Publisher and developer name cleaning, whitespace collapsing, punctuation normalization, and alias deduplication.
- **`internal/metrics/`**:
  - `portfolio.go`: Lorenz curve coordinates, Herfindahl-Hirschman Index (HHI), Top-N concentration shares, effective title count ($N_{\text{eff}} = 1/\sum s_i^2$).
  - `genre.go`: $1/K$ multi-genre proportional allocation ensuring portfolio composition sums to exactly 100.0%.
  - `pricing.go`: Configurable price tier histograms, free-to-play ratios, median/mean pricing.
  - `reviews.go`: Wilson score confidence intervals, positive review fractions, median playtime, peak CCU.
- **`internal/analytics/`**:
  - `outliers.go`: Tukey IQR fence outlier detection ($Q_1 - 1.5\text{IQR}$, $Q_3 + 1.5\text{IQR}$) across ownership, playtime, and peak CCU.
  - `service.go`: Aggregates publisher metrics, summaries, and catalog query filtering.
- **`internal/comparison/comparison.go`**: Multi-publisher comparison matrix (up to 4 publishers) with synchronized scale normalization and coverage warnings.
- **`internal/reports/export.go`**: Streaming CSV and formatted JSON exporter for publishers, catalogs, and dependency reports.

---

### 3.2 Frontend (`frontend/`)

- Built with **React 19**, **TypeScript**, and **Vite**, using **Tailwind CSS v4**.
- Zero heavy third-party charting dependencies: Custom responsive **SVG visualizations** crafted for analytical clarity and fast rendering:
  - `ContributionBarChart.tsx`: Ranked ownership contribution waterfall with interactive title breakdown.
  - `HorizontalBarChart.tsx`: Multi-genre weighted share and raw title comparison.
  - `PriceTierStackedBar.tsx`: Segmented color-coded pricing tier distribution.
  - `TimelineChart.tsx`: Dual-axis release volume and estimated ownership midpoint growth.
  - `OutlierDistributionPlot.tsx`: Statistical Tukey boxplot with scatter outliers and whisker fences.
  - `DeveloperTree.tsx`: Interactive studio collaboration network showing exclusive vs multi-publisher developers.
- Responsive features:
  - **Dashboard**: High-level market catalog totals, concentration tier distribution, top publishers by ownership and catalog size.
  - **Publisher Directory**: Server-side search, filtering by genre, minimum catalog size, and sorting with instant pagination.
  - **Publisher Report**: Complete Portfolio Dependency Report, HHI index, Lorenz curve, genre specialization, pricing tiers, developer network, timeline, and outlier detection.
  - **Catalog Explorer**: Complete game table with search, genre/developer filters, price tiers, quality badges, and game drawer.
  - **Multi-Publisher Compare**: Side-by-side benchmarking of 2-4 publishers with delta indicators.
  - **Data Quality Observatory**: Missing fields diagnostic breakdown, playtest detection, and problematic records inspector.
  - **Dataset Manager**: Hugging Face one-click download, manual Parquet upload, checksum verification, and import job logs.

---

## 4. Concurrency & Performance

- **Ingestion Velocity**: Vectorized DuckDB batch execution imports 124,146 titles, normalizes 63,070 publishers and 75,731 developers, and populates index structures in **6.05 seconds**.
- **Query Latency**: Materialized summary tables and B-tree indices on primary entity identifiers ensure average HTTP endpoint responses under 50ms.
- **Embedded Portability**: The database persists to a single local file (`data/observatory.duckdb`), allowing zero-overhead restarts and persistent offline use.
