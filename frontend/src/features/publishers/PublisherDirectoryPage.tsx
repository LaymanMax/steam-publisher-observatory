import {
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  Download,
  Search,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { api } from "../../services/api";
import type { PublisherSummary } from "../../types";

export const PublisherDirectoryPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const [publishers, setPublishers] = useState<PublisherSummary[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  // Filters from URL
  const query = searchParams.get("q") || "";
  const sortBy = searchParams.get("sort_by") || "catalog_size";
  const sortOrder = searchParams.get("sort_order") || "desc";
  const page = parseInt(searchParams.get("page") || "1", 10);
  const minGames = parseInt(searchParams.get("min_games") || "0", 10);

  const limit = 25;
  const offset = (page - 1) * limit;

  useEffect(() => {
    fetchPublishers();
  }, [query, sortBy, sortOrder, page, minGames]);

  const fetchPublishers = async () => {
    try {
      setLoading(true);
      const res = await api.listPublishers({
        query: query || undefined,
        min_games: minGames > 0 ? minGames : undefined,
        sort_by: sortBy,
        sort_order: sortOrder,
        limit,
        offset,
      });
      setPublishers(res.publishers || []);
      setTotal(res.total || 0);
    } catch (err) {
      console.error("Failed to load publishers:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    const next = new URLSearchParams(searchParams);
    if (val) {
      next.set("q", val);
    } else {
      next.delete("q");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  const handleSort = (col: string) => {
    const next = new URLSearchParams(searchParams);
    if (sortBy === col) {
      next.set("sort_order", sortOrder === "asc" ? "desc" : "asc");
    } else {
      next.set("sort_by", col);
      next.set("sort_order", "desc");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  const handleMinGamesChange = (mg: number) => {
    const next = new URLSearchParams(searchParams);
    if (mg > 0) {
      next.set("min_games", String(mg));
    } else {
      next.delete("min_games");
    }
    next.set("page", "1");
    setSearchParams(next);
  };

  const totalPages = Math.ceil(total / limit) || 1;

  const handlePageChange = (p: number) => {
    if (p < 1 || p > totalPages) return;
    const next = new URLSearchParams(searchParams);
    next.set("page", String(p));
    setSearchParams(next);
  };

  return (
    <div className="container-max py-8 space-y-6">
      {/* Header & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[#1e293b] pb-6">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-white">
            Publisher Directory
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Browse and benchmark {total.toLocaleString()} normalized publishing
            entities.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <a
            href={`/api/exports/publishers.csv?${searchParams.toString()}`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-slate-200 hover:text-white hover:bg-[#1e2c45] transition-standard"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Export CSV</span>
          </a>
        </div>
      </div>

      {/* Filter toolbar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 bg-[#111827] p-3 rounded-lg border border-[#1e293b]">
        {/* Search Input */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
          <input
            type="text"
            placeholder="Search publisher name..."
            value={query}
            onChange={handleSearchChange}
            className="w-full pl-9 pr-3 py-1.5 rounded-md bg-[#090d16] border border-[#1e293b] text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>

        {/* Filter Quick Pills */}
        <div className="flex items-center space-x-2 text-xs">
          <span className="text-slate-500 text-[11px] font-mono mr-1">
            MIN CATALOG:
          </span>
          {[0, 5, 20, 50].map((num) => (
            <button
              key={num}
              onClick={() => handleMinGamesChange(num)}
              className={`px-2.5 py-1 rounded text-xs font-mono transition-standard ${
                minGames === num
                  ? "bg-cyan-500/10 text-cyan-400 border border-cyan-500/30"
                  : "bg-[#090d16] text-slate-400 border border-[#1e293b] hover:text-white"
              }`}
            >
              {num === 0 ? "All" : `${num}+`}
            </button>
          ))}
        </div>
      </div>

      {/* Data Table */}
      <div className="rounded-lg border border-[#1e293b] bg-[#111827] overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs table-compact">
            <thead className="bg-[#090d16] text-slate-400 border-b border-[#1e293b] font-mono text-[11px] select-none">
              <tr>
                <th
                  onClick={() => handleSort("catalog_size")}
                  className="cursor-pointer hover:text-cyan-400 py-3"
                >
                  <div className="flex items-center space-x-1">
                    <span>Publisher</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort("catalog_size")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Catalog
                </th>
                <th
                  onClick={() => handleSort("developer_count")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Studios
                </th>
                <th
                  onClick={() => handleSort("genre_count")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Genres
                </th>
                <th
                  onClick={() => handleSort("catalog_owners_midpoint")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Ownership Midpoint
                </th>
                <th
                  onClick={() => handleSort("top5_share")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Top 5 Share
                </th>
                <th
                  onClick={() => handleSort("hhi")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  HHI Conc.
                </th>
                <th
                  onClick={() => handleSort("ownership_coverage")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Coverage
                </th>
                <th
                  onClick={() => handleSort("avg_quality_score")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Completeness
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e293b]">
              {loading ? (
                <tr>
                  <td
                    colSpan={9}
                    className="py-12 text-center text-slate-500 font-mono"
                  >
                    Loading publishers...
                  </td>
                </tr>
              ) : publishers.length === 0 ? (
                <tr>
                  <td
                    colSpan={9}
                    className="py-12 text-center text-slate-500 font-mono"
                  >
                    No publishers matched your query criteria.
                  </td>
                </tr>
              ) : (
                publishers.map((p) => {
                  let concColor = "text-emerald-400";
                  if (p.top5_share >= 70.0 || p.hhi >= 2500)
                    concColor = "text-rose-400";
                  else if (p.top5_share >= 45.0 || p.hhi >= 1500)
                    concColor = "text-amber-400";

                  return (
                    <tr
                      key={p.publisher_id}
                      className="hover:bg-[#162032] transition-standard group"
                    >
                      <td className="py-2.5 font-medium">
                        <Link
                          to={`/publishers/${p.publisher_id}`}
                          className="text-white group-hover:text-cyan-400 font-semibold flex items-center space-x-1.5"
                        >
                          <span
                            className="truncate max-w-[220px]"
                            title={p.name}
                          >
                            {p.name}
                          </span>
                        </Link>
                        {p.top_title_name && (
                          <div className="text-[10px] text-slate-500 font-mono truncate max-w-[220px]">
                            Top: {p.top_title_name}
                          </div>
                        )}
                      </td>
                      <td className="text-right font-mono text-white font-semibold">
                        {p.catalog_size.toLocaleString()}
                      </td>
                      <td className="text-right font-mono text-slate-300">
                        {p.developer_count.toLocaleString()}
                      </td>
                      <td className="text-right font-mono text-slate-300">
                        {p.genre_count.toLocaleString()}
                      </td>
                      <td className="text-right font-mono text-cyan-400 font-semibold">
                        {p.catalog_owners_midpoint.toLocaleString()}
                      </td>
                      <td
                        className={`text-right font-mono font-bold ${concColor}`}
                      >
                        {p.top5_share.toFixed(1)}%
                      </td>
                      <td className="text-right font-mono text-slate-400">
                        {Math.round(p.hhi).toLocaleString()}
                      </td>
                      <td className="text-right font-mono text-slate-300">
                        {p.ownership_coverage.toFixed(0)}%
                      </td>
                      <td className="text-right font-mono text-slate-400">
                        {p.avg_quality_score.toFixed(0)}%
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className="p-3 border-t border-[#1e293b] bg-[#090d16] flex items-center justify-between text-xs text-slate-400 font-mono">
          <div>
            Showing {total === 0 ? 0 : offset + 1} -{" "}
            {Math.min(offset + limit, total)} of {total.toLocaleString()}{" "}
            publishers
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={() => handlePageChange(page - 1)}
              disabled={page <= 1}
              className="p-1 rounded bg-[#162032] border border-[#23334d] disabled:opacity-40 disabled:cursor-not-allowed hover:bg-[#1f2d47]"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span>
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => handlePageChange(page + 1)}
              disabled={page >= totalPages}
              className="p-1 rounded bg-[#162032] border border-[#23334d] disabled:opacity-40 disabled:cursor-not-allowed hover:bg-[#1f2d47]"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
