# Steam Publisher Observatory

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.8-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![DuckDB](https://img.shields.io/badge/Database-DuckDB-FFF000?logo=duckdb)](https://duckdb.org)

**Steam Publisher Observatory** is an analytical intelligence platform evaluating video game publishers on Steam at the portfolio and corporate level. Designed for game publishers, independent studios, investors, market researchers, and video game industry analysts, it analyzes catalog concentration, studio collaboration networks, multi-genre specialization, pricing structures, and historical release cadences across 124,000+ commercial Steam titles.

---

## Key Highlights & Differentiators

- **Portfolio Dependency Report**: Analyzes catalog reliance on top-performing titles (Top 1, 3, 5, 10 shares, Lorenz curves, Herfindahl-Hirschman Index [HHI], and Effective Title Count $N_{\text{eff}} = 1/\sum s_i^2$).
- **Multi-Genre Proportional Allocation**: Uses $1/K$ weighted attribution for multi-genre titles, ensuring portfolio composition sums to exactly 100.0%.
- **Studio Collaboration Network**: Maps developer credits across publisher catalogs, revealing whether publishers rely on internal studios or diverse external partners.
- **Tukey IQR Outlier Fences**: Non-parametric statistical fences ($Q_1 - 1.5\text{IQR}$, $Q_3 + 1.5\text{IQR}$) detecting hit breakout games and performance anomalies without Gaussian assumptions.
- **Multi-Publisher Benchmarking**: Synchronized side-by-side comparison matrix for up to 4 publishers with delta indicators.
- **Data Quality Observatory**: Transparent accounting of sparse metadata, playtest builds, missing publisher credits, and completeness scores.
- **Embedded Columnar Performance**: Vectorized DuckDB analytical storage ingesting 124,146 titles in **6.05 seconds** with sub-50ms query responses and zero external database dependencies.
- **Single Unified Binary or Vite Dev Mode**: Production Go binary serves both REST APIs and the built React SPA, with hot-reloading Vite dev proxy for local development.

---

## Quick Start

### 1. Prerequisites

- **Go**: 1.24 or later
- **Node.js**: 22 or later
- **pnpm**: 10 or later (`npm install -g pnpm`)

### 2. Build the Application

```bash
# Clone the repository
git clone https://github.com/LaymanMax/steam-publisher-observatory.git
cd steam-publisher-observatory

# Build the React frontend
cd frontend
pnpm install
pnpm build
cd ..

# Build the Go backend binary
cd backend
go build -o bin/server cmd/server/main.go
cd ..
```

### 3. Run the Observatory Server

```bash
# Start the unified backend (serves both API and SPA on port 8080)
./backend/bin/server -port 8080 -db ./data/observatory.duckdb
```

Open **`http://localhost:8080`** in your browser.

If you already have `train-00000-of-00001.parquet` in `./data/`, it is already ingested and indexed into `data/observatory.duckdb` (124,146 titles ready).

---

## Dataset Ingestion

The application uses the public dataset [`FronkonGames/steam-games-dataset`](https://huggingface.co/datasets/FronkonGames/steam-games-dataset).

### Ingest via Web UI

Navigate to **Dataset Manager** (`/dataset`) to:

1. One-click download directly from Hugging Face Hub (with automatic SHA-256 verification).
2. Or drag-and-drop your local `.parquet` file to trigger the ingestion pipeline.

### Ingest via CLI

To ingest from the command line:

```bash
cd backend
go run cmd/ingest/main.go -parquet ../data/train-00000-of-00001.parquet -db ../data/observatory.duckdb
```

---

## Architecture

```text
steam-publisher-observatory/
├── backend/
│   ├── cmd/
│   │   ├── server/          # Go HTTP Server entrypoint
│   │   └── ingest/          # Standalone CLI ingestion tool
│   ├── internal/
│   │   ├── analytics/       # Outliers (Tukey IQR), aggregations, service layer
│   │   ├── api/             # Chi router, handlers, CORS, SPA static server
│   │   ├── comparison/      # Multi-publisher comparison matrix
│   │   ├── config/          # Environment & flag loader
│   │   ├── database/        # Embedded DuckDB schema & connection pool
│   │   ├── dataset/         # Hugging Face metadata discovery & downloader
│   │   ├── importer/        # Vectorized Parquet ingestion & data quality
│   │   ├── jobs/            # Async background import job manager
│   │   ├── metrics/         # Portfolio dependency, HHI, Lorenz, genre weighting
│   │   ├── normalization/   # Entity cleaning, Unicode NFKC, alias resolution
│   │   └── reports/         # CSV streaming & JSON report export
│   └── tests/               # Backend integration test suite
├── frontend/
│   ├── src/
│   │   ├── components/      # SVG charts (Contribution, Timeline, Outlier, Tree)
│   │   ├── features/        # Feature screens (Dashboard, Directory, Report, Compare)
│   │   ├── services/        # Type-safe API client
│   │   ├── types/           # TypeScript data models
│   │   └── test/            # Vitest unit test suite
│   ├── vite.config.ts       # Vite + Tailwind v4 + Proxy setup
│   └── package.json
├── data/                    # Parquet datasets & DuckDB storage (gitignored)
├── docs/                    # Detailed architecture and methodology documentation
│   ├── architecture.md
│   ├── data-dictionary.md
│   ├── metrics-methodology.md
│   └── dataset-import.md
└── README.md
```

---

## Running Automated Tests

### Backend Unit & Integration Tests

```bash
cd backend
go test -v ./...
```

Includes:

- Exact 72.0% top 5 share synthetic dependency calculation test.
- Tukey IQR fence outlier tests.
- Estimated owners range parsing (en-dashes, em-dashes, commas, reversed bounds).
- Entity normalization and Unicode NFKC clean-name tests.
- 24 HTTP API endpoint integration tests.

### Frontend Unit & Component Tests

```bash
cd frontend
pnpm test
```

Includes Vitest and React Testing Library tests for navigation, stat cards, segmented pricing bars, waterfall contribution charts, and API client contracts.

---

## REST API Reference

| Endpoint                                     | Method     | Description                                                             |
| -------------------------------------------- | ---------- | ----------------------------------------------------------------------- |
| `/api/health`                                | `GET`      | Health check and version status.                                        |
| `/api/overview`                              | `GET`      | Platform-wide totals, top publishers, and concentration tiers.          |
| `/api/publishers`                            | `GET`      | Paginated publisher directory with search, genre, and size filters.     |
| `/api/publishers/{id}`                       | `GET`      | Publisher entity details.                                               |
| `/api/publishers/{id}/summary`               | `GET`      | Materialized portfolio summary and concentration metrics.               |
| `/api/publishers/{id}/catalog`               | `GET`      | Paginated game catalog with filters and sort orders.                    |
| `/api/publishers/{id}/dependency`            | `GET`      | Complete Portfolio Dependency Report (Top-N shares, HHI, Lorenz curve). |
| `/api/publishers/{id}/genres`                | `GET`      | Proportional 1/K weighted genre distribution (sums to 100%).            |
| `/api/publishers/{id}/prices`                | `GET`      | Pricing tier distribution, median/mean, free vs paid ratio.             |
| `/api/publishers/{id}/developers`            | `GET`      | Studio collaboration network and developer credits.                     |
| `/api/publishers/{id}/timeline`              | `GET`      | Annual release history and ownership midpoint growth.                   |
| `/api/publishers/{id}/outliers`              | `GET`      | Tukey IQR outlier detection across ownership, playtime, and CCU.        |
| `/api/publishers/{id}/quality`               | `GET`      | Publisher-specific data quality and completeness score.                 |
| `/api/publishers/{id}/report`                | `GET`      | Aggregated corporate dossier combining all analytical dimensions.       |
| `/api/compare?publishers=id1,id2`            | `GET`      | Side-by-side benchmarking of up to 4 publishers.                        |
| `/api/dataset/status`                        | `GET`      | Ingestion status, row counts, and active download jobs.                 |
| `/api/dataset/quality`                       | `GET`      | Global data quality diagnostics and missing field counts.               |
| `/api/dataset/issues`                        | `GET`      | Paginated list of problematic/sparse records.                           |
| `/api/entities/aliases`                      | `GET/POST` | Entity alias consolidation management.                                  |
| `/api/exports/publishers.csv`                | `GET`      | CSV export of publisher directory.                                      |
| `/api/exports/catalog.csv?publisher_id={id}` | `GET`      | CSV export of publisher game catalog.                                   |
| `/api/exports/report.json?publisher_id={id}` | `GET`      | Formatted JSON export of publisher report.                              |

---

## Methodology & Attribution Disclaimers

> **Analytical Disclaimer**: All metrics derived in Steam Publisher Observatory reflect **estimated catalog ownership ranges** and public store metadata collected from Steam. They **do NOT represent audited commercial revenue, net publisher receipts, retail units sold, profit margins, or distinct concurrent account players**.

Read the full methodologies in the documentation:

- [Architecture Guide](./docs/architecture.md)
- [Data Dictionary](./docs/data-dictionary.md)
- [Metrics & Analytical Methodology](./docs/metrics-methodology.md)
- [Dataset Import Guide](./docs/dataset-import.md)

---

## License

This project is licensed under the [MIT License](LICENSE).
The Steam Games Dataset is published by FronkonGames under the MIT license.
Steam and the Steam logo are trademarks of Valve Corporation. This project is not affiliated with Valve Corporation.
