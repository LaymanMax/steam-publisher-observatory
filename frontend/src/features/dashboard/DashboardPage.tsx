import {
  ArrowRight,
  Compass,
  Download,
  Percent,
  TrendingUp,
  Upload,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { StatCard } from "../../components/common/StatCard";
import { api } from "../../services/api";
import type { OverviewMetrics } from "../../types";

export const DashboardPage: React.FC = () => {
  const [data, setData] = useState<OverviewMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadOverview();
  }, []);

  const loadOverview = async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await api.getOverview();
      setData(res);
    } catch (err: any) {
      setError(
        err.message || "Failed to connect to backend observatory service",
      );
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container-max py-12 text-center text-slate-400">
        <div className="inline-block h-8 w-8 animate-spin rounded-full border-2 border-cyan-400 border-t-transparent" />
        <p className="mt-3 text-xs font-mono">
          Connecting to analytical storage engine...
        </p>
      </div>
    );
  }

  // Initial State: No dataset imported yet
  if (!data || data.total_games === 0) {
    return (
      <div className="container-max py-16">
        <div className="max-w-2xl mx-auto text-center space-y-6">
          <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 mb-2">
            <Compass className="h-6 w-6" />
          </div>

          <h1 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Steam Publisher Observatory
          </h1>

          <p className="text-base text-slate-300 leading-relaxed">
            Understand the portfolios behind the games. Steam Publisher
            Observatory analyzes the businesses publishing games on Steam
            instead of focusing exclusively on individual titles.
          </p>

          <div className="p-4 rounded-lg bg-[#111827] border border-[#1e293b] text-left text-xs text-slate-400 space-y-2">
            <div className="font-semibold text-slate-200">
              Public Dataset Requirement
            </div>
            <p>
              This observatory uses the open{" "}
              <strong>FronkonGames/steam-games-dataset</strong> (~192 MB Parquet
              format). Before analyzing publisher portfolios, dependency shares,
              and concentration indexes, import the dataset.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
            <Link
              to="/dataset"
              className="w-full sm:w-auto inline-flex items-center justify-center space-x-2 px-5 py-2.5 rounded-md bg-cyan-500 text-slate-950 hover:bg-cyan-400 font-semibold text-xs transition-standard"
            >
              <Download className="h-4 w-4" />
              <span>Download from Hugging Face</span>
            </Link>
            <Link
              to="/dataset"
              className="w-full sm:w-auto inline-flex items-center justify-center space-x-2 px-5 py-2.5 rounded-md bg-[#162032] border border-[#23334d] text-slate-200 hover:bg-[#1f2d47] font-semibold text-xs transition-standard"
            >
              <Upload className="h-4 w-4" />
              <span>Upload Local Parquet File</span>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container-max py-8 space-y-8">
      {/* Top Banner & Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#1e293b] pb-6">
        <div>
          <h1 className="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
            <span>Portfolio Intelligence Dashboard</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Analyzing commercial publishing catalogs, ownership dependencies,
            and portfolio risk across Steam.
          </p>
        </div>

        <div className="flex items-center space-x-3 text-xs">
          <Link
            to="/publishers"
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-md bg-cyan-500 text-slate-950 hover:bg-cyan-400 font-semibold text-xs transition-standard"
          >
            <span>Explore All Publishers</span>
            <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>
      </div>

      {/* Primary KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        <StatCard
          label="Tracked Games"
          value={data.total_games.toLocaleString()}
          subtext="Indexed commercial titles"
        />
        <StatCard
          label="Publishers"
          value={data.total_publishers.toLocaleString()}
          subtext="Normalized entities"
        />
        <StatCard
          label="Developers"
          value={data.total_developers.toLocaleString()}
          subtext="Observed studio credits"
        />
        <StatCard
          label="Represented Genres"
          value={data.total_genres.toLocaleString()}
          subtext="Distinct genre categories"
        />
        <StatCard
          label="Ownership Coverage"
          value={`${data.ownership_coverage.toFixed(1)}%`}
          subtext="Games with valid ranges"
          tooltip="Percentage of catalog titles with parseable ownership estimate ranges."
          badge={{ text: "Range Aggregated", variant: "info" }}
        />
      </div>

      {/* Analytical Insights Callout */}
      {data.insights && data.insights.length > 0 && (
        <div className="rounded-lg border border-cyan-500/20 bg-gradient-to-r from-cyan-950/20 to-blue-950/20 p-4 space-y-2">
          <div className="text-xs font-semibold text-cyan-400 uppercase font-mono tracking-wider flex items-center">
            <TrendingUp className="h-3.5 w-3.5 mr-1.5" />
            Key Portfolio Observations
          </div>
          <div className="space-y-1 text-xs text-slate-300">
            {data.insights.map((ins, i) => (
              <p key={i} className="flex items-start">
                <span className="text-cyan-400 mr-2 font-mono">•</span>
                <span>{ins}</span>
              </p>
            ))}
          </div>
        </div>
      )}

      {/* Three Comparative Analytical Columns */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 1. Largest Catalogs */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-4 space-y-3">
          <div className="flex items-center justify-between pb-2 border-b border-[#1e293b]">
            <span className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
              Largest Catalogs
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              By Title Count
            </span>
          </div>

          <div className="divide-y divide-[#1e293b]">
            {data.largest_publishers?.map((p) => (
              <Link
                key={p.publisher_id}
                to={`/publishers/${p.publisher_id}`}
                className="py-2.5 flex items-center justify-between group block hover:bg-[#162032] -mx-2 px-2 rounded transition-standard text-xs"
              >
                <div className="truncate pr-2">
                  <div className="font-medium text-slate-200 group-hover:text-cyan-400 truncate">
                    {p.name}
                  </div>
                  <div className="text-[11px] text-slate-500 font-mono">
                    {p.developer_count} developers | {p.genre_count} genres
                  </div>
                </div>
                <div className="text-right font-mono shrink-0">
                  <span className="font-semibold text-white">
                    {p.catalog_size}
                  </span>{" "}
                  <span className="text-slate-500 text-[10px]">titles</span>
                </div>
              </Link>
            ))}
          </div>
        </div>

        {/* 2. Highly Concentrated Portfolios */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-4 space-y-3">
          <div className="flex items-center justify-between pb-2 border-b border-[#1e293b]">
            <span className="text-xs font-semibold text-rose-300 uppercase font-mono tracking-wide flex items-center">
              <Percent className="h-3.5 w-3.5 mr-1 text-rose-400" />
              Highest Top 5 Concentration
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              Min 5 titles
            </span>
          </div>

          <div className="divide-y divide-[#1e293b]">
            {data.concentrated_publishers?.map((p) => (
              <Link
                key={p.publisher_id}
                to={`/publishers/${p.publisher_id}`}
                className="py-2.5 flex items-center justify-between group block hover:bg-[#162032] -mx-2 px-2 rounded transition-standard text-xs"
              >
                <div className="truncate pr-2">
                  <div className="font-medium text-slate-200 group-hover:text-cyan-400 truncate">
                    {p.name}
                  </div>
                  <div className="text-[11px] text-slate-500 font-mono truncate">
                    Top: {p.top_title_name || "N/A"}
                  </div>
                </div>
                <div className="text-right font-mono shrink-0">
                  <div className="font-bold text-rose-400">
                    {p.top5_share.toFixed(1)}%
                  </div>
                  <div className="text-[10px] text-slate-500">
                    HHI: {Math.round(p.hhi)}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </div>

        {/* 3. Diversified Portfolios */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-4 space-y-3">
          <div className="flex items-center justify-between pb-2 border-b border-[#1e293b]">
            <span className="text-xs font-semibold text-emerald-300 uppercase font-mono tracking-wide">
              Most Diversified Portfolios
            </span>
            <span className="text-[11px] text-slate-500 font-mono">
              Lowest HHI (10+ titles)
            </span>
          </div>

          <div className="divide-y divide-[#1e293b]">
            {data.diversified_publishers?.map((p) => (
              <Link
                key={p.publisher_id}
                to={`/publishers/${p.publisher_id}`}
                className="py-2.5 flex items-center justify-between group block hover:bg-[#162032] -mx-2 px-2 rounded transition-standard text-xs"
              >
                <div className="truncate pr-2">
                  <div className="font-medium text-slate-200 group-hover:text-cyan-400 truncate">
                    {p.name}
                  </div>
                  <div className="text-[11px] text-slate-500 font-mono">
                    {p.catalog_size} games |{" "}
                    {p.effective_title_count.toFixed(1)} eff. titles
                  </div>
                </div>
                <div className="text-right font-mono shrink-0">
                  <div className="font-semibold text-emerald-400">
                    HHI {Math.round(p.hhi)}
                  </div>
                  <div className="text-[10px] text-slate-500">
                    Top 5: {p.top5_share.toFixed(1)}%
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
