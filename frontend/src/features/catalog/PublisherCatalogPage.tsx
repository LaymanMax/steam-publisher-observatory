import {
  ArrowLeft,
  ArrowUpDown,
  ChevronLeft,
  ChevronRight,
  Download,
  Search,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { GameDrawer } from "../../components/common/GameDrawer";
import { api } from "../../services/api";
import type { GameCatalogItem, PublisherSummary } from "../../types";

export const PublisherCatalogPage: React.FC = () => {
  const { publisherId } = useParams<{ publisherId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();

  const [summary, setSummary] = useState<PublisherSummary | null>(null);
  const [games, setGames] = useState<GameCatalogItem[]>([]);
  const [total, setTotal] = useState(0);
  const [selectedGame, setSelectedGame] = useState<GameCatalogItem | null>(
    null,
  );
  const [loading, setLoading] = useState(true);

  // Filters from URL
  const query = searchParams.get("q") || "";
  const genre = searchParams.get("genre") || "";
  const priceTier = searchParams.get("price_tier") || "";
  const sortBy = searchParams.get("sort_by") || "owners_midpoint";
  const sortOrder = searchParams.get("sort_order") || "desc";
  const page = parseInt(searchParams.get("page") || "1", 10);
  const excludePlaytest = searchParams.get("exclude_playtest") === "true";

  const limit = 25;
  const offset = (page - 1) * limit;

  useEffect(() => {
    if (publisherId) {
      api
        .getPublisherSummary(publisherId)
        .then(setSummary)
        .catch(() => {});
      fetchCatalog(publisherId);
    }
  }, [
    publisherId,
    query,
    genre,
    priceTier,
    sortBy,
    sortOrder,
    page,
    excludePlaytest,
  ]);

  const fetchCatalog = async (id: string) => {
    try {
      setLoading(true);
      const res = await api.getPublisherCatalog(id, {
        query: query || undefined,
        genre: genre || undefined,
        price_tier: priceTier || undefined,
        exclude_playtest: excludePlaytest,
        sort_by: sortBy,
        sort_order: sortOrder,
        limit,
        offset,
      });
      setGames(res.games || []);
      setTotal(res.total || 0);
    } catch (err) {
      console.error("Failed to load catalog:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    const next = new URLSearchParams(searchParams);
    if (val) next.set("q", val);
    else next.delete("q");
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

  const handleExcludePlaytestToggle = () => {
    const next = new URLSearchParams(searchParams);
    if (!excludePlaytest) {
      next.set("exclude_playtest", "true");
    } else {
      next.delete("exclude_playtest");
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
      {/* Top Breadcrumb & Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-[#1e293b] pb-6">
        <div>
          <div className="flex items-center space-x-2 text-xs text-slate-400 mb-1">
            <Link
              to="/publishers"
              className="hover:text-cyan-400 transition-standard"
            >
              Publishers
            </Link>
            <span>/</span>
            <Link
              to={`/publishers/${publisherId}`}
              className="hover:text-cyan-400 transition-standard"
            >
              {summary?.name || "Publisher Report"}
            </Link>
            <span>/</span>
            <span className="text-white">Catalog</span>
          </div>

          <h1 className="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
            <span>{summary?.name} — Catalog Explorer</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Inspecting {total.toLocaleString()} commercial titles with granular
            engagement, pricing, and ownership data.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <Link
            to={`/publishers/${publisherId}`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-slate-200 hover:text-white hover:bg-[#1e2d47] transition-standard"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>Portfolio Report</span>
          </Link>
          <a
            href={`/api/exports/catalog.csv?publisher_id=${publisherId}&${searchParams.toString()}`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-cyan-500 text-slate-950 hover:bg-cyan-400 font-semibold text-xs transition-standard"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Export CSV</span>
          </a>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 bg-[#111827] p-3 rounded-lg border border-[#1e293b]">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
          <input
            type="text"
            placeholder="Search game title..."
            value={query}
            onChange={handleSearchChange}
            className="w-full pl-9 pr-3 py-1.5 rounded-md bg-[#090d16] border border-[#1e293b] text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>

        {/* Playtest Toggle */}
        <label className="flex items-center space-x-2 text-xs text-slate-300 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={excludePlaytest}
            onChange={handleExcludePlaytestToggle}
            className="rounded border-[#1e293b] text-cyan-500 focus:ring-0 bg-[#090d16]"
          />
          <span>Exclude suspected playtests</span>
        </label>
      </div>

      {/* Catalog Table */}
      <div className="rounded-lg border border-[#1e293b] bg-[#111827] overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs table-compact">
            <thead className="bg-[#090d16] text-slate-400 border-b border-[#1e293b] font-mono text-[11px] select-none">
              <tr>
                <th
                  onClick={() => handleSort("name")}
                  className="cursor-pointer hover:text-cyan-400 py-3"
                >
                  <div className="flex items-center space-x-1">
                    <span>Title</span>
                    <ArrowUpDown className="h-3 w-3" />
                  </div>
                </th>
                <th
                  onClick={() => handleSort("release_date")}
                  className="cursor-pointer hover:text-cyan-400"
                >
                  Release
                </th>
                <th
                  onClick={() => handleSort("price")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Price
                </th>
                <th
                  onClick={() => handleSort("owners_midpoint")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Owners Range
                </th>
                <th
                  onClick={() => handleSort("owners_midpoint")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Midpoint
                </th>
                <th
                  onClick={() => handleSort("total_reviews")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Reviews
                </th>
                <th
                  onClick={() => handleSort("positive_share")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Pos. %
                </th>
                <th
                  onClick={() => handleSort("peak_ccu")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Peak CCU
                </th>
                <th
                  onClick={() => handleSort("quality_score")}
                  className="cursor-pointer hover:text-cyan-400 text-right"
                >
                  Quality
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
                    Loading catalog titles...
                  </td>
                </tr>
              ) : games.length === 0 ? (
                <tr>
                  <td
                    colSpan={9}
                    className="py-12 text-center text-slate-500 font-mono"
                  >
                    No games found matching criteria.
                  </td>
                </tr>
              ) : (
                games.map((g) => (
                  <tr
                    key={g.appid}
                    onClick={() => setSelectedGame(g)}
                    className="hover:bg-[#162032] transition-standard cursor-pointer group"
                  >
                    <td className="py-2.5 font-medium">
                      <div className="flex items-center space-x-2">
                        <span className="text-white group-hover:text-cyan-400 font-semibold truncate max-w-[260px]">
                          {g.name}
                        </span>
                        {g.is_playtest && (
                          <span className="px-1.5 py-0.2 rounded text-[9px] font-mono bg-rose-500/10 text-rose-400 border border-rose-500/20">
                            Playtest
                          </span>
                        )}
                      </div>
                      <div className="text-[10px] text-slate-500 font-mono flex items-center space-x-2">
                        <span>AppID: {g.appid}</span>
                        {g.developers?.length > 0 && (
                          <span>• Dev: {g.developers.join(", ")}</span>
                        )}
                      </div>
                    </td>
                    <td className="font-mono text-slate-300">
                      {g.release_date ||
                        (g.release_year ? String(g.release_year) : "Unknown")}
                    </td>
                    <td className="text-right font-mono text-white">
                      {g.price !== undefined ? (
                        g.price === 0 ? (
                          <span className="text-emerald-400 font-semibold">
                            Free
                          </span>
                        ) : (
                          `$${g.price.toFixed(2)}`
                        )
                      ) : (
                        <span className="text-slate-500">Missing</span>
                      )}
                    </td>
                    <td className="text-right font-mono text-slate-400 text-[11px]">
                      {g.owners_raw || "Missing"}
                    </td>
                    <td className="text-right font-mono text-cyan-400 font-semibold">
                      {g.owners_midpoint !== undefined
                        ? g.owners_midpoint.toLocaleString()
                        : "N/A"}
                    </td>
                    <td className="text-right font-mono text-slate-300">
                      {g.total_reviews.toLocaleString()}
                    </td>
                    <td className="text-right font-mono text-emerald-400 font-medium">
                      {g.positive_share !== undefined
                        ? `${g.positive_share.toFixed(0)}%`
                        : "—"}
                    </td>
                    <td className="text-right font-mono text-slate-300">
                      {g.peak_ccu.toLocaleString()}
                    </td>
                    <td className="text-right font-mono text-slate-400">
                      {g.quality_score.toFixed(0)}%
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className="p-3 border-t border-[#1e293b] bg-[#090d16] flex items-center justify-between text-xs text-slate-400 font-mono">
          <div>
            Showing {total === 0 ? 0 : offset + 1} -{" "}
            {Math.min(offset + limit, total)} of {total.toLocaleString()} games
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

      {/* Game Detail Drawer */}
      <GameDrawer game={selectedGame} onClose={() => setSelectedGame(null)} />
    </div>
  );
};
