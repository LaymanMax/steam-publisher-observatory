import {
  AlertTriangle,
  Download,
  GitCompare,
  Info,
  ShieldCheck,
  Table,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ContributionBarChart } from "../../components/charts/ContributionBarChart";
import { DeveloperTree } from "../../components/charts/DeveloperTree";
import { HorizontalBarChart } from "../../components/charts/HorizontalBarChart";
import { OutlierDistributionPlot } from "../../components/charts/OutlierDistributionPlot";
import { PriceTierStackedBar } from "../../components/charts/PriceTierStackedBar";
import { TimelineChart } from "../../components/charts/TimelineChart";
import { StatCard } from "../../components/common/StatCard";
import { api } from "../../services/api";
import type {
  DeveloperAnalysisReport,
  GenreAnalysisReport,
  OutlierAnalysisReport,
  PortfolioDependencyReport,
  PricingReport,
  PublisherQualityReport,
  PublisherSummary,
  TimelineReport,
} from "../../types";

export const PublisherReportPage: React.FC = () => {
  const { publisherId } = useParams<{ publisherId: string }>();

  const [summary, setSummary] = useState<PublisherSummary | null>(null);
  const [dependency, setDependency] =
    useState<PortfolioDependencyReport | null>(null);
  const [genres, setGenres] = useState<GenreAnalysisReport | null>(null);
  const [pricing, setPricing] = useState<PricingReport | null>(null);
  const [developers, setDevelopers] = useState<DeveloperAnalysisReport | null>(
    null,
  );
  const [timeline, setTimeline] = useState<TimelineReport | null>(null);
  const [outliers, setOutliers] = useState<OutlierAnalysisReport | null>(null);
  const [quality, setQuality] = useState<PublisherQualityReport | null>(null);

  const [selectedOutlierMetric, setSelectedOutlierMetric] =
    useState("owners_midpoint");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (publisherId) {
      loadReport(publisherId);
    }
  }, [publisherId]);

  useEffect(() => {
    if (publisherId) {
      api
        .getPublisherOutliers(publisherId, selectedOutlierMetric)
        .then(setOutliers)
        .catch(() => {});
    }
  }, [publisherId, selectedOutlierMetric]);

  const loadReport = async (id: string) => {
    try {
      setLoading(true);
      setError(null);

      const [sum, dep, gen, pri, dev, tim, out, qua] = await Promise.all([
        api.getPublisherSummary(id),
        api.getPublisherDependency(id),
        api.getPublisherGenres(id),
        api.getPublisherPrices(id),
        api.getPublisherDevelopers(id),
        api.getPublisherTimeline(id),
        api.getPublisherOutliers(id, selectedOutlierMetric),
        api.getPublisherQuality(id),
      ]);

      setSummary(sum);
      setDependency(dep);
      setGenres(gen);
      setPricing(pri);
      setDevelopers(dev);
      setTimeline(tim);
      setOutliers(out);
      setQuality(qua);
    } catch (err: any) {
      setError(err.message || "Failed to load publisher portfolio report");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="container-max py-16 text-center text-slate-400">
        <div className="inline-block h-8 w-8 animate-spin rounded-full border-2 border-cyan-400 border-t-transparent" />
        <p className="mt-3 text-xs font-mono">
          Synthesizing publisher portfolio analytics...
        </p>
      </div>
    );
  }

  if (error || !summary) {
    return (
      <div className="container-max py-16 text-center text-slate-400 space-y-4">
        <AlertTriangle className="h-10 w-10 text-amber-400 mx-auto" />
        <h2 className="text-lg font-bold text-white">
          Publisher Report Unavailable
        </h2>
        <p className="text-xs text-slate-400">
          {error || "Publisher not found."}
        </p>
        <Link
          to="/publishers"
          className="inline-flex px-4 py-1.5 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-cyan-400 hover:bg-[#1e2d47]"
        >
          Return to Publisher Directory
        </Link>
      </div>
    );
  }

  const concBadgeVariant =
    dependency?.concentration_level === "High"
      ? "danger"
      : dependency?.concentration_level === "Moderate"
        ? "warning"
        : "success";

  return (
    <div className="container-max py-8 space-y-8">
      {/* Executive Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 border-b border-[#1e293b] pb-6">
        <div>
          <div className="flex items-center space-x-3">
            <h1 className="text-2xl font-bold tracking-tight text-white">
              {summary.name}
            </h1>
            {dependency && (
              <span
                className={`px-2 py-0.5 rounded text-[11px] font-mono font-semibold uppercase tracking-wider border ${
                  concBadgeVariant === "danger"
                    ? "bg-rose-500/10 text-rose-400 border-rose-500/20"
                    : concBadgeVariant === "warning"
                      ? "bg-amber-500/10 text-amber-400 border-amber-500/20"
                      : "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                }`}
              >
                {dependency.concentration_level} Concentration
              </span>
            )}
          </div>
          <p className="text-xs text-slate-400 mt-1 font-mono">
            Publisher Identity Key: {summary.publisher_id} | Active Catalog:{" "}
            {summary.catalog_size} games
          </p>
        </div>

        {/* Action toolbar */}
        <div className="flex flex-wrap items-center gap-2">
          <Link
            to={`/publishers/${publisherId}/catalog`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-slate-200 hover:text-white hover:bg-[#1f2d47] transition-standard"
          >
            <Table className="h-3.5 w-3.5 text-cyan-400" />
            <span>Catalog Explorer</span>
          </Link>
          <Link
            to={`/compare?publishers=${publisherId}`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] text-xs font-medium text-slate-200 hover:text-white hover:bg-[#1f2d47] transition-standard"
          >
            <GitCompare className="h-3.5 w-3.5 text-blue-400" />
            <span>Benchmark Publisher</span>
          </Link>
          <a
            href={`/api/exports/report.json?publisher_id=${publisherId}`}
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-cyan-500 text-slate-950 font-semibold text-xs hover:bg-cyan-400 transition-standard"
          >
            <Download className="h-3.5 w-3.5" />
            <span>Export JSON Report</span>
          </a>
        </div>
      </div>

      {/* A. Executive Overview Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <StatCard
          label="Catalog Size"
          value={summary.catalog_size.toLocaleString()}
          subtext={`${summary.valid_owners_count} with owner estimates`}
        />
        <StatCard
          label="Studios Credited"
          value={summary.developer_count.toLocaleString()}
          subtext="Unique development studios"
        />
        <StatCard
          label="Represented Genres"
          value={summary.genre_count.toLocaleString()}
          subtext={
            summary.dominant_genre
              ? `Dominant: ${summary.dominant_genre}`
              : "Genres"
          }
        />
        <StatCard
          label="Release Horizon"
          value={
            summary.earliest_year && summary.latest_year
              ? `${summary.earliest_year} - ${summary.latest_year}`
              : "N/A"
          }
          subtext="Historical catalog span"
        />
        <StatCard
          label="Ownership Midpoint"
          value={summary.catalog_owners_midpoint.toLocaleString()}
          subtext={`Range: ${(summary.catalog_owners_lower / 1000).toFixed(0)}k - ${(summary.catalog_owners_upper / 1000).toFixed(0)}k`}
          tooltip="Aggregated midpoint of per-game estimated ownership ranges. Illustrative metric only; not unique customers or sales."
        />
        <StatCard
          label="Completeness"
          value={`${summary.avg_quality_score.toFixed(0)}%`}
          subtext={`${summary.ownership_coverage.toFixed(0)}% owner coverage`}
          badge={{ text: "Data Quality", variant: "info" }}
        />
      </div>

      {/* G. PORTFOLIO DEPENDENCY REPORT (Hero Section) */}
      {dependency && (
        <section className="rounded-xl border border-cyan-500/30 bg-gradient-to-br from-[#0c1322] to-[#111928] p-6 shadow-md space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-[#1e293b] pb-4">
            <div>
              <div className="flex items-center space-x-2">
                <span className="h-2 w-2 rounded-full bg-cyan-400 animate-pulse" />
                <h2 className="text-base font-bold text-white uppercase font-mono tracking-wider">
                  Portfolio Dependency Report
                </h2>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Evaluates how heavily catalog ownership estimates are
                concentrated in a small subset of titles.
              </p>
            </div>
            <div className="font-mono text-xs text-cyan-400 font-semibold bg-cyan-500/10 px-2.5 py-1 rounded border border-cyan-500/20">
              Concentration Index (HHI): {Math.round(dependency.scaled_hhi)}
            </div>
          </div>

          {/* Headline Metric Banner */}
          <div className="p-4 rounded-lg bg-[#080d17] border border-[#1e293b] space-y-2">
            <div className="text-base sm:text-lg font-bold text-white leading-snug">
              {dependency.headline_metric}
            </div>
            <p className="text-xs text-slate-300 leading-relaxed">
              {dependency.interpretation}
            </p>
          </div>

          {/* Key Concentration KPIs */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Top 1 Title Share
              </div>
              <div className="text-xl font-bold font-mono text-white mt-1">
                {dependency.top1_share.toFixed(1)}%
              </div>
            </div>
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Top 3 Titles Share
              </div>
              <div className="text-xl font-bold font-mono text-white mt-1">
                {dependency.top3_share.toFixed(1)}%
              </div>
            </div>
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Top 5 Titles Share
              </div>
              <div className="text-xl font-bold font-mono text-cyan-400 mt-1">
                {dependency.top5_share.toFixed(1)}%
              </div>
            </div>
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Top 10 Titles Share
              </div>
              <div className="text-xl font-bold font-mono text-white mt-1">
                {dependency.top10_share.toFixed(1)}%
              </div>
            </div>
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Remaining Tail Share
              </div>
              <div className="text-xl font-bold font-mono text-slate-300 mt-1">
                {dependency.remaining_share.toFixed(1)}%
              </div>
            </div>
            <div className="p-3 rounded-lg bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 font-mono uppercase">
                Effective Title Count
              </div>
              <div
                className="text-xl font-bold font-mono text-emerald-400 mt-1"
                title="1 / unscaled_HHI: equivalent number of equally-sized titles in catalog"
              >
                {dependency.effective_title_count.toFixed(1)}
              </div>
            </div>
          </div>

          {/* Ranked Contribution Stack & Cumulative Curve */}
          <div className="space-y-3 pt-2">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-medium text-white">
                Individual Title Contributions to Catalog Midpoint
              </span>
              <span className="text-[11px] font-mono text-slate-500">
                Ranked descending by ownership midpoint
              </span>
            </div>
            <ContributionBarChart
              titles={dependency.ranked_titles}
              maxDisplay={10}
            />
          </div>

          {/* Methodology & Limitations Popover/Notice */}
          <div className="p-3 rounded-md bg-[#080d17]/80 border border-[#1e293b] text-xs text-slate-400 flex items-start space-x-2.5">
            <Info className="h-4 w-4 text-cyan-400 shrink-0 mt-0.5" />
            <div className="space-y-1">
              <strong className="text-slate-300">
                Methodology & Essential Limitations:
              </strong>{" "}
              <span>{dependency.methodology_limitations}</span> Never equate
              ownership concentration with revenue dependency, profitability,
              sales, or unique active players.
            </div>
          </div>
        </section>
      )}

      {/* Two-Column Grid: Portfolio Composition & Pricing */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* B1. Genre Composition */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
                Genre Distribution & Diversification
              </h3>
              <p className="text-[11px] text-slate-400">
                Multi-genre titles weighted equally (1/K) so portfolio
                composition sums to 100%.
              </p>
            </div>
            {genres && (
              <span className="text-xs font-mono text-slate-400">
                {genres.distinct_genre_count} genres
              </span>
            )}
          </div>
          {genres && (
            <HorizontalBarChart items={genres.genres} maxDisplay={8} />
          )}
        </div>

        {/* B2. Pricing Tiers */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
                Pricing Tiers & Distribution
              </h3>
              <p className="text-[11px] text-slate-400">
                Median: ${pricing?.median_price.toFixed(2) || "0.00"} | Mean: $
                {pricing?.avg_price.toFixed(2) || "0.00"} (USD)
              </p>
            </div>
            {pricing && (
              <span className="text-xs font-mono text-emerald-400">
                {pricing.free_share.toFixed(1)}% Free titles
              </span>
            )}
          </div>
          {pricing && <PriceTierStackedBar tiers={pricing.tiers} />}
        </div>
      </div>

      {/* C. Catalog Timeline */}
      {timeline && (
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
                Historical Release Timeline
              </h3>
              <p className="text-[11px] text-slate-400">
                Annual title volume and ownership midpoint trajectory over time.
              </p>
            </div>
            <span className="text-xs font-mono text-slate-400">
              {timeline.years.length} active release years
            </span>
          </div>
          <TimelineChart years={timeline.years} />
        </div>
      )}

      {/* D. Developer Collaboration Relationships */}
      {developers && (
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
                Developer Collaboration Patterns
              </h3>
              <p className="text-[11px] text-slate-400">
                Observed game credits associating development studios with
                titles published by this catalog.
              </p>
            </div>
            <span className="text-xs font-mono text-cyan-400">
              {developers.total_developers} studios credited
            </span>
          </div>
          <DeveloperTree report={developers} />
        </div>
      )}

      {/* Outlier Detection Section */}
      {outliers && (
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
                Explainable Outlier Detection (Tukey IQR)
              </h3>
              <p className="text-[11px] text-slate-400">
                Statistical anomalies within this publisher's own portfolio
                based on interquartile boundaries.
              </p>
            </div>
          </div>
          <OutlierDistributionPlot
            report={outliers}
            selectedMetric={selectedOutlierMetric}
            onSelectMetric={setSelectedOutlierMetric}
          />
        </div>
      )}

      {/* H. Data Quality & Metadata Incompleteness */}
      {quality && (
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide flex items-center">
                <ShieldCheck className="h-4 w-4 mr-1.5 text-cyan-400" />
                Data Quality & Observational Limitations
              </h3>
              <p className="text-[11px] text-slate-400">
                Incomplete or missing source metadata that impacts report
                precision.
              </p>
            </div>
            <span className="text-xs font-mono text-white font-semibold">
              Quality Score: {quality.completeness_score.toFixed(0)}%
            </span>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 text-xs">
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Missing Owners
              </div>
              <div className="text-lg font-mono font-bold text-white mt-1">
                {quality.missing_owners_count}
              </div>
            </div>
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Missing Prices
              </div>
              <div className="text-lg font-mono font-bold text-white mt-1">
                {quality.missing_price_count}
              </div>
            </div>
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Missing Dates
              </div>
              <div className="text-lg font-mono font-bold text-white mt-1">
                {quality.missing_date_count}
              </div>
            </div>
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Missing Genres
              </div>
              <div className="text-lg font-mono font-bold text-white mt-1">
                {quality.missing_genre_count}
              </div>
            </div>
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Missing Studios
              </div>
              <div className="text-lg font-mono font-bold text-white mt-1">
                {quality.missing_developer_count}
              </div>
            </div>
            <div className="p-3 rounded bg-[#0d131f] border border-[#1e293b]">
              <div className="text-[10px] text-slate-500 uppercase font-mono">
                Suspected Playtests
              </div>
              <div className="text-lg font-mono font-bold text-rose-400 mt-1">
                {quality.playtest_count}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
