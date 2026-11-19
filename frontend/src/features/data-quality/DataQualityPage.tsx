import { ExternalLink, ShieldAlert } from "lucide-react";
import React, { useEffect, useState } from "react";
import { StatCard } from "../../components/common/StatCard";
import { api } from "../../services/api";
import type {
  GlobalDataQualityDiagnostics,
  ProblematicRecord,
} from "../../types";

export const DataQualityPage: React.FC = () => {
  const [diagnostics, setDiagnostics] =
    useState<GlobalDataQualityDiagnostics | null>(null);
  const [records, setRecords] = useState<ProblematicRecord[]>([]);
  const [totalRecords, setTotalRecords] = useState(0);
  const [selectedIssue, setSelectedIssue] = useState("playtest");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDiagnostics();
  }, []);

  useEffect(() => {
    loadIssues(selectedIssue);
  }, [selectedIssue]);

  const loadDiagnostics = async () => {
    try {
      setLoading(true);
      const diag = await api.getDatasetQuality();
      setDiagnostics(diag);
    } catch (err) {
      console.error("Failed to load quality diagnostics:", err);
    } finally {
      setLoading(false);
    }
  };

  const loadIssues = async (issue: string) => {
    try {
      const res = await api.getDatasetIssues(issue, 50, 0);
      setRecords(res.records || []);
      setTotalRecords(res.total || 0);
    } catch (err) {
      console.error("Failed to load issues:", err);
    }
  };

  const issuesList = [
    { id: "playtest", label: "Suspected Playtests" },
    { id: "sparse", label: "Sparse Records" },
    { id: "missing_owners", label: "Missing Owners" },
    { id: "missing_publisher", label: "Missing Publisher" },
    { id: "missing_developer", label: "Missing Studio" },
    { id: "missing_price", label: "Missing Price" },
    { id: "missing_date", label: "Invalid Date" },
  ];

  if (loading) {
    return (
      <div className="container-max py-16 text-center text-slate-400 font-mono text-xs">
        Loading data quality diagnostics...
      </div>
    );
  }

  return (
    <div className="container-max py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-[#1e293b] pb-6">
        <h1 className="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
          <ShieldAlert className="h-5 w-5 text-cyan-400" />
          <span>Data Quality Observatory</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Diagnostics, completeness scoring, and heuristic anomaly flags across
          the underlying Steam games dataset.
        </p>
      </div>

      {/* Global Diagnostics Grid */}
      {diagnostics && (
        <div className="space-y-4">
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            <StatCard
              label="Total Rows"
              value={diagnostics.total_rows.toLocaleString()}
              subtext={`${diagnostics.unique_appids.toLocaleString()} unique AppIDs`}
            />
            <StatCard
              label="Completeness"
              value={`${diagnostics.average_completeness.toFixed(1)}%`}
              subtext="Catalog average score"
              badge={{ text: "Global Quality", variant: "info" }}
            />
            <StatCard
              label="Missing Publishers"
              value={diagnostics.games_without_publishers.toLocaleString()}
              subtext={`${((diagnostics.games_without_publishers / diagnostics.total_rows) * 100).toFixed(1)}% of rows`}
            />
            <StatCard
              label="Missing Studios"
              value={diagnostics.games_without_developers.toLocaleString()}
              subtext={`${((diagnostics.games_without_developers / diagnostics.total_rows) * 100).toFixed(1)}% of rows`}
            />
            <StatCard
              label="Suspected Playtests"
              value={diagnostics.suspected_playtests.toLocaleString()}
              subtext="Flagged by heuristic"
              badge={{ text: "Heuristic", variant: "warning" }}
            />
            <StatCard
              label="Sparse Records"
              value={diagnostics.sparse_records.toLocaleString()}
              subtext="Minimal metadata present"
              badge={{ text: "Sparse", variant: "danger" }}
            />
          </div>

          {/* Additional details */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="text-[11px] text-slate-400">
                Missing Release Dates
              </span>
              <div className="text-base font-mono font-bold text-white">
                {diagnostics.missing_release_dates.toLocaleString()}
              </div>
            </div>
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="text-[11px] text-slate-400">Missing Prices</span>
              <div className="text-base font-mono font-bold text-white">
                {diagnostics.missing_prices.toLocaleString()}
              </div>
            </div>
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="text-[11px] text-slate-400">Missing Genres</span>
              <div className="text-base font-mono font-bold text-white">
                {diagnostics.missing_genres.toLocaleString()}
              </div>
            </div>
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="text-[11px] text-slate-400">
                Missing Ownership Ranges
              </span>
              <div className="text-base font-mono font-bold text-white">
                {diagnostics.missing_owners.toLocaleString()}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Heuristic Explanations */}
      <div className="p-4 rounded-lg bg-[#111827] border border-[#1e293b] space-y-2 text-xs">
        <h3 className="font-semibold text-slate-200">
          Playtest & Sparse Record Policy
        </h3>
        <p className="text-slate-400 leading-relaxed">
          The dataset includes non-commercial playtests (e.g.,{" "}
          <em>Black Dragon Mage Playtest</em>), unreleased tests, and zero-owner
          sparse entries. Rather than silently discarding these rows, the
          observatory preserves them, calculates a transparent completeness
          score, and flags them for optional exclusion.
        </p>
      </div>

      {/* Problematic Records Explorer */}
      <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-[#1e293b]">
          <div>
            <h3 className="text-xs font-semibold text-white uppercase font-mono tracking-wide">
              Problematic Record Inspector
            </h3>
            <p className="text-[11px] text-slate-400">
              Showing sample records flagged with specific metadata quality
              defects.
            </p>
          </div>

          <div className="flex flex-wrap gap-1.5">
            {issuesList.map((issue) => (
              <button
                key={issue.id}
                onClick={() => setSelectedIssue(issue.id)}
                className={`px-2.5 py-1 rounded text-xs font-medium transition-standard ${
                  selectedIssue === issue.id
                    ? "bg-[#162032] text-cyan-400 border border-[#23334d]"
                    : "bg-[#090d16] text-slate-400 border border-[#1e293b] hover:text-white"
                }`}
              >
                {issue.label}
              </button>
            ))}
          </div>
        </div>

        {/* Records Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs table-compact">
            <thead className="bg-[#090d16] text-slate-400 border-b border-[#1e293b] font-mono text-[11px]">
              <tr>
                <th>AppID</th>
                <th>Title</th>
                <th>Completeness Score</th>
                <th>Quality Flags</th>
                <th className="text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e293b]">
              {records.length === 0 ? (
                <tr>
                  <td
                    colSpan={5}
                    className="py-8 text-center text-slate-500 font-mono"
                  >
                    No records found for this issue flag.
                  </td>
                </tr>
              ) : (
                records.map((r) => (
                  <tr
                    key={r.appid}
                    className="hover:bg-[#162032] transition-standard"
                  >
                    <td className="font-mono text-slate-400">{r.appid}</td>
                    <td className="font-medium text-white truncate max-w-[280px]">
                      {r.name || "<Unnamed Title>"}
                    </td>
                    <td className="font-mono text-cyan-400 font-semibold">
                      {r.quality_score.toFixed(0)}%
                    </td>
                    <td className="text-[11px] font-mono text-slate-400">
                      {r.flags}
                    </td>
                    <td className="text-right">
                      <a
                        href={`https://store.steampowered.com/app/${r.appid}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center space-x-1 text-cyan-400 hover:text-cyan-300 text-[11px]"
                      >
                        <span>Steam</span>
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
