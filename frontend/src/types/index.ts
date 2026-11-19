export interface OverviewMetrics {
  total_games: number;
  total_publishers: number;
  total_developers: number;
  total_genres: number;
  ownership_coverage: number;
  dataset_freshness?: string;
  source?: string;
  largest_publishers?: PublisherSummary[];
  concentrated_publishers?: PublisherSummary[];
  diversified_publishers?: PublisherSummary[];
  insights?: string[];
}

export interface PublisherSummary {
  publisher_id: string;
  name: string;
  catalog_size: number;
  developer_count: number;
  genre_count: number;
  earliest_year?: number;
  latest_year?: number;
  valid_owners_count: number;
  ownership_coverage: number;
  catalog_owners_lower: number;
  catalog_owners_midpoint: number;
  catalog_owners_upper: number;
  top_title_name?: string;
  top_title_midpoint: number;
  top1_share: number;
  top3_share: number;
  top5_share: number;
  top10_share: number;
  remaining_share: number;
  hhi: number;
  effective_title_count: number;
  free_title_count: number;
  free_title_share: number;
  median_price: number;
  avg_price: number;
  total_positive_reviews: number;
  total_negative_reviews: number;
  total_reviews: number;
  avg_positive_share: number;
  median_playtime: number;
  avg_peak_ccu: number;
  max_peak_ccu: number;
  dominant_genre?: string;
  genre_concentration: number;
  avg_quality_score: number;
}

export interface GameCatalogItem {
  appid: string;
  name: string;
  developers: string[];
  genres: string[];
  release_date?: string;
  release_year?: number;
  price?: number;
  price_tier: string;
  owners_raw?: string;
  owners_lower?: number;
  owners_upper?: number;
  owners_midpoint?: number;
  positive: number;
  negative: number;
  total_reviews: number;
  positive_share?: number;
  peak_ccu: number;
  median_playtime: number;
  header_image?: string;
  short_description?: string;
  quality_score: number;
  is_playtest: boolean;
  is_sparse: boolean;
}

export interface GameOwnershipContribution {
  appid: string;
  name: string;
  owners_lower: number;
  owners_upper: number;
  owners_midpoint: number;
  share_of_midpoint: number;
  cumulative_share: number;
  rank: number;
}

export interface PortfolioDependencyReport {
  publisher_id: string;
  publisher_name: string;
  catalog_size: number;
  valid_owners_count: number;
  ownership_coverage: number;
  catalog_owners_lower: number;
  catalog_owners_midpoint: number;
  catalog_owners_upper: number;
  top1_share: number;
  top3_share: number;
  top5_share: number;
  top10_share: number;
  remaining_share: number;
  scaled_hhi: number;
  unscaled_hhi: number;
  effective_title_count: number;
  ranked_titles: GameOwnershipContribution[];
  headline_metric: string;
  interpretation: string;
  concentration_level: "High" | "Moderate" | "Diversified" | string;
  methodology_limitations: string;
}

export interface GenreItem {
  genre: string;
  game_count: number;
  weighted_weight: number;
  weighted_share: number;
  ownership_midpoint: number;
}

export interface GenreYearTrend {
  year: number;
  total_games: number;
  genre_counts: Record<string, number>;
  genre_shares: Record<string, number>;
}

export interface GenreAnalysisReport {
  publisher_id: string;
  total_catalog_games: number;
  games_with_genres: number;
  genre_coverage: number;
  distinct_genre_count: number;
  dominant_genre: string;
  dominant_genre_share: number;
  genre_concentration: number;
  genres: GenreItem[];
  timeline?: GenreYearTrend[];
}

export interface PriceTierItem {
  tier_name: string;
  min_price: number;
  max_price: number;
  game_count: number;
  share: number;
}

export interface PricingReport {
  publisher_id: string;
  total_games: number;
  free_count: number;
  free_share: number;
  paid_count: number;
  paid_share: number;
  missing_count: number;
  missing_share: number;
  min_price: number;
  max_price: number;
  avg_price: number;
  median_price: number;
  tiers: PriceTierItem[];
  currency: string;
}

export interface DeveloperRelationshipItem {
  developer_id: string;
  developer_name: string;
  associated_games: number;
  share_of_catalog: number;
  owners_midpoint: number;
  shared_publishers: number;
  games: Array<{
    appid: string;
    name: string;
    release_year?: number;
    midpoint: number;
  }>;
}

export interface DeveloperAnalysisReport {
  publisher_id: string;
  total_developers: number;
  single_title_developers: number;
  multi_title_developers: number;
  developer_concentration: number;
  top_developer_name: string;
  top_developer_share: number;
  developers: DeveloperRelationshipItem[];
  notice: string;
}

export interface TimelineYearEntry {
  year: number;
  game_count: number;
  owners_midpoint: number;
  free_count: number;
  paid_count: number;
  developer_count: number;
  total_reviews: number;
  avg_positive_share: number;
}

export interface TimelineReport {
  publisher_id: string;
  earliest_year?: number;
  latest_year?: number;
  unknown_date_games: number;
  years: TimelineYearEntry[];
}

export interface OutlierItem {
  appid: string;
  name: string;
  metric_name: string;
  metric_value: number;
  outlier_type: "high" | "low";
  q1: number;
  median: number;
  q3: number;
  iqr: number;
  upper_fence: number;
  lower_fence: number;
  rank_in_metric: number;
  total_count: number;
}

export interface OutlierAnalysisReport {
  metric_name: string;
  sample_count: number;
  q1: number;
  median: number;
  q3: number;
  iqr: number;
  upper_fence: number;
  lower_fence: number;
  outliers: OutlierItem[];
  methodology: string;
  has_outliers: boolean;
  explanation: string;
}

export interface PublisherQualityReport {
  publisher_id: string;
  total_games: number;
  completeness_score: number;
  missing_owners_count: number;
  missing_price_count: number;
  missing_date_count: number;
  missing_genre_count: number;
  missing_developer_count: number;
  playtest_count: number;
  sparse_count: number;
}

export interface GlobalDataQualityDiagnostics {
  total_rows: number;
  unique_appids: number;
  duplicate_appids: number;
  games_with_publishers: number;
  games_without_publishers: number;
  games_with_developers: number;
  games_without_developers: number;
  missing_genres: number;
  missing_release_dates: number;
  missing_prices: number;
  missing_owners: number;
  suspected_playtests: number;
  sparse_records: number;
  average_completeness: number;
}

export interface ProblematicRecord {
  appid: string;
  name: string;
  issue: string;
  quality_score: number;
  flags: string;
}

export interface PublisherComparisonProfile {
  publisher_id: string;
  name: string;
  catalog_size: number;
  developer_count: number;
  genre_count: number;
  free_title_share: number;
  median_price: number;
  ownership_coverage: number;
  catalog_owners_midpoint: number;
  top5_share: number;
  top10_share: number;
  hhi: number;
  effective_title_count: number;
  total_reviews: number;
  positive_share: number;
  median_playtime: number;
  avg_peak_ccu: number;
  max_peak_ccu: number;
  genres: GenreItem[];
  pricing_tiers: PriceTierItem[];
  top_titles: Array<{
    name: string;
    midpoint: number;
    share: number;
  }>;
}

export interface PublisherComparisonResult {
  publishers: PublisherComparisonProfile[];
  metric_ranges: Record<string, { min: number; max: number }>;
  warnings: string[];
}

export interface ImportJob {
  id: string;
  type: string;
  stage: string;
  progress: number;
  message: string;
  status: "pending" | "running" | "completed" | "failed";
  error?: string;
  started_at: string;
  completed_at?: string;
}

export interface EntityAliasRecord {
  id: string;
  entity_type: string;
  source_name: string;
  canonical_id: string;
  canonical_name: string;
  created_at: string;
  method: string;
}
