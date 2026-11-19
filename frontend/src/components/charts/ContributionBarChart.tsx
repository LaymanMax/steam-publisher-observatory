import React, { useState } from "react";
import type { GameOwnershipContribution } from "../../types";

interface ContributionBarChartProps {
  titles: GameOwnershipContribution[];
  maxDisplay?: number;
}

export const ContributionBarChart: React.FC<ContributionBarChartProps> = ({
  titles,
  maxDisplay = 10,
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  if (!titles || titles.length === 0) {
    return (
      <div className="flex h-48 items-center justify-center text-xs text-slate-500 border border-dashed border-[#1e293b] rounded">
        No valid ownership contributions to chart.
      </div>
    );
  }

  const displayed = titles.slice(0, maxDisplay);

  return (
    <div className="space-y-4">
      {/* Top Titles Ranked Bar Stack */}
      <div className="space-y-2">
        {displayed.map((t, idx) => {
          const isTop5 = idx < 5;
          const isHovered = hoveredIndex === idx;

          return (
            <div
              key={t.appid}
              className={`p-2 rounded border transition-standard cursor-pointer ${
                isHovered
                  ? "border-cyan-500/50 bg-[#162032]"
                  : "border-[#1e293b] bg-[#0d131f] hover:border-slate-700"
              }`}
              onMouseEnter={() => setHoveredIndex(idx)}
              onMouseLeave={() => setHoveredIndex(null)}
            >
              <div className="flex items-center justify-between text-xs mb-1">
                <div className="flex items-center space-x-2 truncate">
                  <span className="font-mono text-[11px] text-slate-500 w-5">
                    #{t.rank}
                  </span>
                  <span
                    className="font-medium text-slate-200 truncate"
                    title={t.name}
                  >
                    {t.name}
                  </span>
                </div>
                <div className="flex items-center space-x-3 font-mono text-[11px] shrink-0">
                  <span className="text-slate-400">
                    {t.owners_midpoint.toLocaleString()} est. midpoint
                  </span>
                  <span
                    className={`font-semibold ${isTop5 ? "text-cyan-400" : "text-slate-300"}`}
                  >
                    {t.share_of_midpoint.toFixed(1)}%
                  </span>
                  <span className="text-slate-500 text-[10px]">
                    (cum: {t.cumulative_share.toFixed(1)}%)
                  </span>
                </div>
              </div>

              {/* Share bar */}
              <div className="h-1.5 w-full bg-[#1e293b] rounded-full overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${
                    isTop5
                      ? "bg-gradient-to-r from-cyan-500 to-blue-500"
                      : "bg-slate-500"
                  }`}
                  style={{
                    width: `${Math.min(100, Math.max(1, t.share_of_midpoint))}%`,
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>

      {titles.length > maxDisplay && (
        <div className="text-center text-xs text-slate-500 pt-1">
          + {titles.length - maxDisplay} additional titles in catalog remaining
          tail
        </div>
      )}
    </div>
  );
};
