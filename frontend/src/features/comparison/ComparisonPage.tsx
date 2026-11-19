import {
  AlertTriangle,
  ArrowRight,
  GitCompare,
  Plus,
  Search,
  X,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { api } from "../../services/api";
import type { PublisherComparisonResult, PublisherSummary } from "../../types";

export const ComparisonPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const [comparison, setComparison] =
    useState<PublisherComparisonResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Search input for adding a publisher
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<PublisherSummary[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const rawPubParam = searchParams.get("publishers") || "";
  const selectedIDs = rawPubParam ? rawPubParam.split(",").filter(Boolean) : [];

  useEffect(() => {
    if (selectedIDs.length > 0) {
      loadComparison(selectedIDs);
    } else {
      setComparison(null);
    }
  }, [rawPubParam]);

  const loadComparison = async (ids: string[]) => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.comparePublishers(ids);
      setComparison(res);
    } catch (err: any) {
      setError(err.message || "Failed to generate comparison");
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async (val: string) => {
    setSearchQuery(val);
    if (!val.trim()) {
      setSearchResults([]);
      return;
    }
    try {
      setIsSearching(true);
      const res = await api.listPublishers({ query: val, limit: 5 });
      setSearchResults(res.publishers || []);
    } catch {
      // ignore
    } finally {
      setIsSearching(false);
    }
  };

  const addPublisher = (id: string) => {
    if (selectedIDs.includes(id)) return;
    if (selectedIDs.length >= 4) return;
    const nextIDs = [...selectedIDs, id];
    const next = new URLSearchParams(searchParams);
    next.set("publishers", nextIDs.join(","));
    setSearchParams(next);
    setSearchQuery("");
    setSearchResults([]);
  };

  const removePublisher = (id: string) => {
    const nextIDs = selectedIDs.filter((i) => i !== id);
    const next = new URLSearchParams(searchParams);
    if (nextIDs.length > 0) {
      next.set("publishers", nextIDs.join(","));
    } else {
      next.delete("publishers");
    }
    setSearchParams(next);
  };

  return (
    <div className="container-max py-8 space-y-6">
      {/* Top Header */}
      <div className="border-b border-[#1e293b] pb-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
            <GitCompare className="h-5 w-5 text-cyan-400" />
            <span>Publisher Benchmark & Comparison</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Compare portfolio dependencies, pricing tiers, and studio networks
            across up to 4 publishers with synchronized scales.
          </p>
        </div>

        {/* Selected Publisher Pills */}
        <div className="flex flex-wrap items-center gap-2">
          {comparison?.publishers.map((p) => (
            <span
              key={p.publisher_id}
              className="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-white"
            >
              <span>{p.name}</span>
              <button
                onClick={() => removePublisher(p.publisher_id)}
                className="text-slate-400 hover:text-rose-400 transition-standard"
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      </div>

      {/* Add Publisher Search Bar (if < 4 selected) */}
      {selectedIDs.length < 4 && (
        <div className="relative max-w-lg bg-[#111827] p-3 rounded-lg border border-[#1e293b]">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
            <input
              type="text"
              placeholder={`Search to add publisher (${selectedIDs.length}/4 selected)...`}
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 rounded bg-[#090d16] border border-[#1e293b] text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* Autocomplete dropdown */}
          {searchResults.length > 0 && (
            <div className="absolute left-0 right-0 top-full mt-1 bg-[#0d131f] border border-[#1e293b] rounded-md shadow-2xl z-30 divide-y divide-[#1e293b] max-h-60 overflow-y-auto">
              {searchResults.map((p) => (
                <div
                  key={p.publisher_id}
                  onClick={() => addPublisher(p.publisher_id)}
                  className="p-2.5 hover:bg-[#162032] cursor-pointer flex items-center justify-between text-xs transition-standard"
                >
                  <div>
                    <span className="font-medium text-white">{p.name}</span>
                    <span className="text-[11px] font-mono text-slate-500 ml-2">
                      ({p.catalog_size} games)
                    </span>
                  </div>
                  <Plus className="h-4 w-4 text-cyan-400" />
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Warnings Banner if coverage varies */}
      {comparison?.warnings && comparison.warnings.length > 0 && (
        <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-200 space-y-1">
          {comparison.warnings.map((w, idx) => (
            <div key={idx} className="flex items-start space-x-2">
              <AlertTriangle className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
              <span>{w}</span>
            </div>
          ))}
        </div>
      )}

      {loading && (
        <div className="py-16 text-center text-slate-400 font-mono text-xs">
          Computing multi-portfolio comparative metrics...
        </div>
      )}

      {/* Empty State when no publishers selected */}
      {!loading && (!comparison || comparison.publishers.length === 0) && (
        <div className="py-16 text-center text-slate-400 space-y-3 bg-[#111827] rounded-lg border border-dashed border-[#1e293b] p-8">
          <GitCompare className="h-8 w-8 text-slate-500 mx-auto" />
          <h2 className="text-base font-bold text-white">
            No Publishers Selected for Comparison
          </h2>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Use the search box above to add up to 4 publishers and compare their
            catalog size, portfolio concentration, and pricing models side by
            side.
          </p>
        </div>
      )}

      {/* Side-by-Side Comparison Grid */}
      {!loading && comparison && comparison.publishers.length > 0 && (
        <div className="space-y-6">
          {/* Aligned Cards Header */}
          <div
            className={`grid grid-cols-1 md:grid-cols-${comparison.publishers.length} gap-4`}
          >
            {comparison.publishers.map((p) => (
              <div
                key={p.publisher_id}
                className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4 shadow-sm"
              >
                <div className="flex items-start justify-between border-b border-[#1e293b] pb-3">
                  <div>
                    <h2
                      className="text-base font-bold text-white truncate max-w-[200px]"
                      title={p.name}
                    >
                      {p.name}
                    </h2>
                    <span className="text-[11px] font-mono text-slate-500">
                      ID: {p.publisher_id}
                    </span>
                  </div>
                  <Link
                    to={`/publishers/${p.publisher_id}`}
                    className="p-1 rounded text-cyan-400 hover:text-cyan-300 hover:bg-[#162032] transition-standard text-xs flex items-center space-x-1"
                    title="Open full publisher report"
                  >
                    <span>Report</span>
                    <ArrowRight className="h-3.5 w-3.5" />
                  </Link>
                </div>

                {/* Metrics Table */}
                <div className="space-y-2.5 text-xs font-mono divide-y divide-[#1e293b]/60">
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Catalog Size:</span>
                    <span className="text-white font-semibold">
                      {p.catalog_size} titles
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Studios:</span>
                    <span className="text-slate-200">{p.developer_count}</span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Genres:</span>
                    <span className="text-slate-200">{p.genre_count}</span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Ownership Midpoint:</span>
                    <span className="text-cyan-400 font-bold">
                      {p.catalog_owners_midpoint.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Coverage:</span>
                    <span className="text-slate-200">
                      {p.ownership_coverage.toFixed(0)}%
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Top 5 Share:</span>
                    <span
                      className={`font-bold ${p.top5_share >= 70 ? "text-rose-400" : "text-slate-200"}`}
                    >
                      {p.top5_share.toFixed(1)}%
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">HHI Concentration:</span>
                    <span className="text-slate-200">{Math.round(p.hhi)}</span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">
                      Effective Title Count:
                    </span>
                    <span className="text-emerald-400 font-semibold">
                      {p.effective_title_count.toFixed(1)}
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Median Price:</span>
                    <span className="text-slate-200">
                      ${p.median_price.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Free Titles:</span>
                    <span className="text-slate-200">
                      {p.free_title_share.toFixed(1)}%
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">Total Reviews:</span>
                    <span className="text-slate-200">
                      {p.total_reviews.toLocaleString()}
                    </span>
                  </div>
                  <div className="flex justify-between pt-1">
                    <span className="text-slate-400">
                      Positive Review Share:
                    </span>
                    <span className="text-emerald-400 font-semibold">
                      {p.positive_share.toFixed(1)}%
                    </span>
                  </div>
                </div>

                {/* Top Titles Preview */}
                <div className="pt-2 border-t border-[#1e293b] space-y-1.5">
                  <span className="text-[10px] font-mono uppercase text-slate-500">
                    Largest Titles
                  </span>
                  {p.top_titles?.slice(0, 3).map((t, idx) => (
                    <div
                      key={idx}
                      className="flex justify-between text-[11px] truncate"
                    >
                      <span className="text-slate-300 truncate max-w-[150px]">
                        {t.name}
                      </span>
                      <span className="font-mono text-cyan-400 shrink-0">
                        {t.share.toFixed(1)}%
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
