import type {
  DeveloperAnalysisReport,
  EntityAliasRecord,
  GameCatalogItem,
  GenreAnalysisReport,
  GlobalDataQualityDiagnostics,
  ImportJob,
  OutlierAnalysisReport,
  OverviewMetrics,
  PortfolioDependencyReport,
  PricingReport,
  ProblematicRecord,
  PublisherComparisonResult,
  PublisherQualityReport,
  PublisherSummary,
  TimelineReport,
} from "../types";

const BASE_URL = "/api";

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, options);
  if (!res.ok) {
    let errMsg = `Request failed: ${res.status} ${res.statusText}`;
    try {
      const errBody = await res.json();
      if (errBody && errBody.error) {
        errMsg = errBody.error;
      }
    } catch {
      // fallback
    }
    throw new Error(errMsg);
  }
  return res.json();
}

export const api = {
  // Overview
  getOverview: () => fetchJSON<OverviewMetrics>(`${BASE_URL}/overview`),

  // Dataset
  getDatasetStatus: () =>
    fetchJSON<{
      total_games: number;
      total_publishers: number;
      total_developers: number;
      ownership_coverage: number;
      dataset_freshness?: string;
      source?: string;
      active_job?: ImportJob;
    }>(`${BASE_URL}/dataset/status`),

  startImport: () =>
    fetchJSON<ImportJob>(`${BASE_URL}/dataset/import`, { method: "POST" }),

  uploadParquet: async (file: File): Promise<ImportJob> => {
    const formData = new FormData();
    formData.append("file", file);
    return fetchJSON<ImportJob>(`${BASE_URL}/dataset/upload`, {
      method: "POST",
      body: formData,
    });
  },

  cancelImport: () =>
    fetchJSON<{ cancelled: boolean; job_id: string }>(
      `${BASE_URL}/dataset/cancel`,
      {
        method: "POST",
      },
    ),

  getDatasetSchema: () =>
    fetchJSON<{ row_count: number; columns: string[] }>(
      `${BASE_URL}/dataset/schema`,
    ),

  getDatasetQuality: () =>
    fetchJSON<GlobalDataQualityDiagnostics>(`${BASE_URL}/dataset/quality`),

  getDatasetIssues: (issue?: string, limit = 50, offset = 0) => {
    const p = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    if (issue) p.set("issue", issue);
    return fetchJSON<{
      records: ProblematicRecord[];
      total: number;
      limit: number;
      offset: number;
    }>(`${BASE_URL}/dataset/issues?${p.toString()}`);
  },

  // Publishers
  listPublishers: (params?: {
    query?: string;
    min_games?: number;
    genre?: string;
    sort_by?: string;
    sort_order?: string;
    limit?: number;
    offset?: number;
  }) => {
    const p = new URLSearchParams();
    if (params?.query) p.set("query", params.query);
    if (params?.min_games) p.set("min_games", String(params.min_games));
    if (params?.genre) p.set("genre", params.genre);
    if (params?.sort_by) p.set("sort_by", params.sort_by);
    if (params?.sort_order) p.set("sort_order", params.sort_order);
    if (params?.limit) p.set("limit", String(params.limit));
    if (params?.offset) p.set("offset", String(params.offset));
    return fetchJSON<{
      publishers: PublisherSummary[];
      total: number;
      limit: number;
      offset: number;
    }>(`${BASE_URL}/publishers?${p.toString()}`);
  },

  getPublisherSummary: (id: string) =>
    fetchJSON<PublisherSummary>(`${BASE_URL}/publishers/${id}/summary`),

  getPublisherCatalog: (
    id: string,
    params?: {
      query?: string;
      genre?: string;
      developer_id?: string;
      price_tier?: string;
      min_year?: number;
      max_year?: number;
      min_reviews?: number;
      exclude_playtest?: boolean;
      sort_by?: string;
      sort_order?: string;
      limit?: number;
      offset?: number;
    },
  ) => {
    const p = new URLSearchParams();
    if (params?.query) p.set("query", params.query);
    if (params?.genre) p.set("genre", params.genre);
    if (params?.developer_id) p.set("developer_id", params.developer_id);
    if (params?.price_tier) p.set("price_tier", params.price_tier);
    if (params?.min_year) p.set("min_year", String(params.min_year));
    if (params?.max_year) p.set("max_year", String(params.max_year));
    if (params?.min_reviews) p.set("min_reviews", String(params.min_reviews));
    if (params?.exclude_playtest) p.set("exclude_playtest", "true");
    if (params?.sort_by) p.set("sort_by", params.sort_by);
    if (params?.sort_order) p.set("sort_order", params.sort_order);
    if (params?.limit) p.set("limit", String(params.limit));
    if (params?.offset) p.set("offset", String(params.offset));
    return fetchJSON<{
      games: GameCatalogItem[];
      total: number;
      limit: number;
      offset: number;
    }>(`${BASE_URL}/publishers/${id}/catalog?${p.toString()}`);
  },

  getPublisherDependency: (id: string) =>
    fetchJSON<PortfolioDependencyReport>(
      `${BASE_URL}/publishers/${id}/dependency`,
    ),

  getPublisherGenres: (id: string) =>
    fetchJSON<GenreAnalysisReport>(`${BASE_URL}/publishers/${id}/genres`),

  getPublisherPrices: (id: string) =>
    fetchJSON<PricingReport>(`${BASE_URL}/publishers/${id}/prices`),

  getPublisherDevelopers: (id: string) =>
    fetchJSON<DeveloperAnalysisReport>(
      `${BASE_URL}/publishers/${id}/developers`,
    ),

  getPublisherTimeline: (id: string) =>
    fetchJSON<TimelineReport>(`${BASE_URL}/publishers/${id}/timeline`),

  getPublisherOutliers: (id: string, metric?: string) => {
    const p = metric ? `?metric=${metric}` : "";
    return fetchJSON<OutlierAnalysisReport>(
      `${BASE_URL}/publishers/${id}/outliers${p}`,
    );
  },

  getPublisherQuality: (id: string) =>
    fetchJSON<PublisherQualityReport>(`${BASE_URL}/publishers/${id}/quality`),

  // Comparison
  comparePublishers: (ids: string[]) => {
    const p = new URLSearchParams({ publishers: ids.join(",") });
    return fetchJSON<PublisherComparisonResult>(
      `${BASE_URL}/compare?${p.toString()}`,
    );
  },

  // Entities & Aliases
  listDevelopers: (params?: {
    query?: string;
    limit?: number;
    offset?: number;
  }) => {
    const p = new URLSearchParams();
    if (params?.query) p.set("query", params.query);
    if (params?.limit) p.set("limit", String(params.limit));
    if (params?.offset) p.set("offset", String(params.offset));
    return fetchJSON<{
      developers: Array<{ id: string; name: string; game_count: number }>;
      total: number;
      limit: number;
      offset: number;
    }>(`${BASE_URL}/entities/developers?${p.toString()}`);
  },

  listAliases: () =>
    fetchJSON<{ aliases: EntityAliasRecord[] }>(`${BASE_URL}/entities/aliases`),

  createAlias: (payload: {
    entity_type: string;
    source_name: string;
    canonical_id: string;
  }) =>
    fetchJSON<{ success: boolean; message: string }>(
      `${BASE_URL}/entities/aliases`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      },
    ),

  deleteAlias: (id: string) =>
    fetchJSON<{ success: boolean; message: string }>(
      `${BASE_URL}/entities/aliases/${id}`,
      { method: "DELETE" },
    ),

  listPotentialAliases: (limit = 20) =>
    fetchJSON<{
      potential_aliases: Array<{
        entity_a: PublisherSummary;
        entity_b: PublisherSummary;
        similarity: number;
        reason: string;
      }>;
    }>(`${BASE_URL}/entities/potential-aliases?limit=${limit}`),
};
