# Data Dictionary: Steam Publisher Observatory

This document defines the schema, data types, field derivations, and quality rules implemented in the Steam Publisher Observatory analytical database (`data/observatory.duckdb`).

---

## 1. Table: `games`

Represents individual titles indexed from the Steam games dataset.

| Column                 | Type      | Nullable | Description                                                                                    |
| ---------------------- | --------- | -------- | ---------------------------------------------------------------------------------------------- |
| `appid`                | `BIGINT`  | No       | Steam Application ID (Primary Key).                                                            |
| `name`                 | `VARCHAR` | No       | Title name as reported by the source.                                                          |
| `release_date`         | `DATE`    | Yes      | Parsed release date (ISO format `YYYY-MM-DD`).                                                 |
| `release_year`         | `INTEGER` | Yes      | Extracted year of release for temporal analytics.                                              |
| `price`                | `DOUBLE`  | No       | Base USD price (0.00 for Free-to-play).                                                        |
| `price_tier`           | `VARCHAR` | No       | Categorical tier: `Free`, `Very Low`, `Low`, `Mid-range`, `Premium`, `High-priced`, `Missing`. |
| `estimated_owners_raw` | `VARCHAR` | Yes      | Raw string from dataset (e.g. `"20000 - 50000"`).                                              |
| `owners_lower`         | `BIGINT`  | No       | Parsed lower bound of estimated ownership range.                                               |
| `owners_upper`         | `BIGINT`  | No       | Parsed upper bound of estimated ownership range.                                               |
| `owners_midpoint`      | `BIGINT`  | No       | Arithmetic midpoint `(owners_lower + owners_upper) / 2`.                                       |
| `peak_ccu`             | `INTEGER` | No       | Historical peak concurrent users recorded on Steam.                                            |
| `positive`             | `INTEGER` | No       | Total count of positive user reviews.                                                          |
| `negative`             | `INTEGER` | No       | Total count of negative user reviews.                                                          |
| `total_reviews`        | `INTEGER` | No       | Computed sum: `positive + negative`.                                                           |
| `positive_share`       | `DOUBLE`  | No       | Computed percentage: `positive / (positive + negative) * 100`.                                 |
| `recommendations`      | `INTEGER` | No       | Steam recommendation count.                                                                    |
| `median_playtime`      | `INTEGER` | No       | Median playtime forever in minutes.                                                            |
| `average_playtime`     | `INTEGER` | No       | Average playtime forever in minutes.                                                           |
| `metacritic_score`     | `INTEGER` | No       | Metacritic score (0 to 100, 0 = unrated).                                                      |
| `user_score`           | `INTEGER` | No       | Steam user score metadata attribute.                                                           |
| `required_age`         | `INTEGER` | No       | Age gate restriction (0 = all ages).                                                           |
| `dlc_count`            | `INTEGER` | No       | Number of associated DLC packages.                                                             |
| `achievements`         | `INTEGER` | No       | Total Steam achievements available.                                                            |
| `windows`              | `BOOLEAN` | No       | Windows OS compatibility.                                                                      |
| `mac`                  | `BOOLEAN` | No       | macOS compatibility.                                                                           |
| `linux`                | `BOOLEAN` | No       | Linux / SteamOS compatibility.                                                                 |
| `header_image`         | `VARCHAR` | Yes      | CDN URL to official Steam store header graphic.                                                |
| `website`              | `VARCHAR` | Yes      | Developer/publisher external website URL.                                                      |
| `support_url`          | `VARCHAR` | Yes      | Customer support URL.                                                                          |
| `short_description`    | `VARCHAR` | Yes      | Steam store summary paragraph.                                                                 |
| `quality_score`        | `DOUBLE`  | No       | Completeness score (0 to 100) assessing metadata fidelity.                                     |
| `is_playtest`          | `BOOLEAN` | No       | Flag identifying pre-release/playtest beta builds.                                             |
| `is_sparse`            | `BOOLEAN` | No       | Flag identifying records lacking descriptions, price, and genres.                              |

---

## 2. Table: `publishers`

Represents distinct publishing corporate entities.

| Column                | Type      | Nullable | Description                                                        |
| --------------------- | --------- | -------- | ------------------------------------------------------------------ |
| `publisher_id`        | `VARCHAR` | No       | Deterministic MD5 hash of canonical normalized name (Primary Key). |
| `display_name`        | `VARCHAR` | No       | Canonical display name.                                            |
| `canonical_key`       | `VARCHAR` | No       | Lowercase, whitespace-stripped lookup key.                         |
| `catalog_size`        | `INTEGER` | No       | Total number of published titles.                                  |
| `first_release_year`  | `INTEGER` | Yes      | Year of earliest published title.                                  |
| `latest_release_year` | `INTEGER` | Yes      | Year of most recent published title.                               |

---

## 3. Table: `developers`

Represents development studios credited on Steam titles.

| Column          | Type      | Nullable | Description                                                        |
| --------------- | --------- | -------- | ------------------------------------------------------------------ |
| `developer_id`  | `VARCHAR` | No       | Deterministic MD5 hash of canonical normalized name (Primary Key). |
| `display_name`  | `VARCHAR` | No       | Canonical studio display name.                                     |
| `canonical_key` | `VARCHAR` | No       | Lowercase, whitespace-stripped lookup key.                         |
| `game_count`    | `INTEGER` | No       | Total titles credited to this studio.                              |

---

## 4. Relationship Tables

### `game_publishers`

Many-to-many link between titles and publishing entities.

- `appid` (`BIGINT`, Foreign Key `games.appid`)
- `publisher_id` (`VARCHAR`, Foreign Key `publishers.publisher_id`)

### `game_developers`

Many-to-many link between titles and development studios.

- `appid` (`BIGINT`, Foreign Key `games.appid`)
- `developer_id` (`VARCHAR`, Foreign Key `developers.developer_id`)

### `game_genres`

Relational mapping of genre classifications per title.

- `appid` (`BIGINT`, Foreign Key `games.appid`)
- `genre` (`VARCHAR`, e.g. `"Action"`, `"Strategy"`, `"RPG"`, `"Indie"`)

---

## 5. Table: `publisher_summaries`

Materialized analytical summary caching computed portfolio metrics for instant search and rendering.

| Column                    | Type      | Description                                         |
| ------------------------- | --------- | --------------------------------------------------- |
| `publisher_id`            | `VARCHAR` | Primary Key.                                        |
| `name`                    | `VARCHAR` | Publisher display name.                             |
| `catalog_size`            | `INTEGER` | Total published games count.                        |
| `developer_count`         | `INTEGER` | Distinct development partners.                      |
| `genre_count`             | `INTEGER` | Distinct genres spanned by portfolio.               |
| `catalog_owners_midpoint` | `BIGINT`  | Sum of title ownership midpoints across catalog.    |
| `top1_share`              | `DOUBLE`  | Ownership percentage concentrated in #1 title.      |
| `top3_share`              | `DOUBLE`  | Ownership percentage concentrated in top 3 titles.  |
| `top5_share`              | `DOUBLE`  | Ownership percentage concentrated in top 5 titles.  |
| `top10_share`             | `DOUBLE`  | Ownership percentage concentrated in top 10 titles. |
| `hhi`                     | `DOUBLE`  | Herfindahl-Hirschman Index (0 to 10,000).           |
| `effective_title_count`   | `DOUBLE`  | Inverted concentration metric $1 / \sum s_i^2$.     |
| `median_price`            | `DOUBLE`  | Portfolio median base price ($ USD).                |
| `avg_positive_share`      | `DOUBLE`  | Portfolio review sentiment average.                 |
| `avg_quality_score`       | `DOUBLE`  | Mean metadata completeness score across catalog.    |

---

## 6. Table: `entity_aliases`

Stores user-curated and auto-detected identity consolidations (e.g. mapping `"Valve Corporation"` to `"Valve"`).

- `id` (`VARCHAR`, UUID)
- `entity_type` (`VARCHAR`, `'publisher'` or `'developer'`)
- `source_name` (`VARCHAR`, Variant string found in raw data)
- `canonical_id` (`VARCHAR`, Target entity identifier)
- `created_at` (`TIMESTAMP`)
