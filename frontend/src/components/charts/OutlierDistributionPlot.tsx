import { AlertCircle, CheckCircle2 } from "lucide-react";
import React from "react";
import type { OutlierAnalysisReport } from "../../types";

interface OutlierDistributionPlotProps {
  report: OutlierAnalysisReport;
  selectedMetric: string;
  onSelectMetric: (metric: string) => void;
}

export const OutlierDistributionPlot: React.FC<
  OutlierDistributionPlotProps
> = ({ report, selectedMetric, onSelectMetric }) => {
  const metricOptions = [
    { id: "owners_midpoint", label: "Ownership Midpoint" },
    { id: "positive_reviews", label: "Positive Reviews" },
    { id: "total_reviews", label: "Total Reviews" },
    { id: "peak_ccu", label: "Peak CCU" },
    { id: "recommendations", label: "Recommendations" },
    { id: "median_playtime", label: "Median Playtime" },
    { id: "price", label: "Price" },
  ];

  return (
    <div className="space-y-4">
      {/* Metric Selector Tabs */}
      <div className="flex flex-wrap gap-1.5 p-1 bg-[#090d16] rounded-md border border-[#1e293b]">
        {metricOptions.map((opt) => (
          <button
            key={opt.id}
            onClick={() => onSelectMetric(opt.id)}
            className={`px-2.5 py-1 rounded text-xs font-medium transition-standard ${
              selectedMetric === opt.id
                ? "bg-[#162032] text-cyan-400 border border-[#2b3d5b]"
                : "text-slate-400 hover:text-white hover:bg-[#111827]"
            }`}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* IQR Boundary Summary Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-2 text-xs">
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            Q1 (25th %)
          </div>
          <div className="text-sm font-mono font-bold text-slate-300 mt-0.5">
            {report.q1.toLocaleString(undefined, { maximumFractionDigits: 1 })}
          </div>
        </div>
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            Median (50th %)
          </div>
          <div className="text-sm font-mono font-bold text-white mt-0.5">
            {report.median.toLocaleString(undefined, {
              maximumFractionDigits: 1,
            })}
          </div>
        </div>
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            Q3 (75th %)
          </div>
          <div className="text-sm font-mono font-bold text-slate-300 mt-0.5">
            {report.q3.toLocaleString(undefined, { maximumFractionDigits: 1 })}
          </div>
        </div>
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            IQR Range
          </div>
          <div className="text-sm font-mono font-bold text-cyan-400 mt-0.5">
            {report.iqr.toLocaleString(undefined, { maximumFractionDigits: 1 })}
          </div>
        </div>
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            Upper Fence (Q3+1.5*IQR)
          </div>
          <div className="text-sm font-mono font-bold text-rose-400 mt-0.5">
            {report.upper_fence.toLocaleString(undefined, {
              maximumFractionDigits: 1,
            })}
          </div>
        </div>
        <div className="p-2.5 rounded bg-[#0d131f] border border-[#1e293b]">
          <div className="text-[10px] text-slate-500 uppercase font-mono">
            Sample Size
          </div>
          <div className="text-sm font-mono font-bold text-slate-300 mt-0.5">
            {report.sample_count} valid
          </div>
        </div>
      </div>

      {/* Explanation Banner */}
      <div
        className={`p-3 rounded-md border text-xs flex items-start space-x-2.5 ${
          report.has_outliers
            ? "bg-amber-500/10 border-amber-500/20 text-amber-200"
            : "bg-[#111827] border-[#1e293b] text-slate-300"
        }`}
      >
        {report.has_outliers ? (
          <AlertCircle className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
        ) : (
          <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0 mt-0.5" />
        )}
        <div className="space-y-1">
          <p className="font-medium">{report.explanation}</p>
          <p className="text-[11px] text-slate-400">{report.methodology}</p>
        </div>
      </div>

      {/* Outlier Titles List */}
      {report.has_outliers && (
        <div className="rounded-md border border-[#1e293b] overflow-hidden">
          <table className="w-full text-left text-xs table-compact">
            <thead className="bg-[#090d16] text-slate-400 border-b border-[#1e293b] font-mono text-[11px]">
              <tr>
                <th>Title</th>
                <th>Observed Value</th>
                <th>Upper Threshold</th>
                <th>Delta vs Fence</th>
                <th>Position</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1e293b] bg-[#111827]">
              {report.outliers.map((o) => {
                const diff = o.metric_value - o.upper_fence;
                return (
                  <tr
                    key={o.appid}
                    className="hover:bg-[#162032] transition-standard"
                  >
                    <td className="font-medium text-white">{o.name}</td>
                    <td className="font-mono text-rose-400 font-semibold">
                      {o.metric_value.toLocaleString(undefined, {
                        maximumFractionDigits: 1,
                      })}
                    </td>
                    <td className="font-mono text-slate-400">
                      {o.upper_fence.toLocaleString(undefined, {
                        maximumFractionDigits: 1,
                      })}
                    </td>
                    <td className="font-mono text-amber-400">
                      +
                      {diff.toLocaleString(undefined, {
                        maximumFractionDigits: 1,
                      })}
                    </td>
                    <td className="font-mono text-slate-400 text-[11px]">
                      Rank #{o.rank_in_metric} of {o.total_count}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
