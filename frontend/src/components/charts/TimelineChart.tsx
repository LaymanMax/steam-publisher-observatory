import React, { useState } from "react";
import type { TimelineYearEntry } from "../../types";

interface TimelineChartProps {
  years: TimelineYearEntry[];
}

export const TimelineChart: React.FC<TimelineChartProps> = ({ years }) => {
  const [hoveredYear, setHoveredYear] = useState<TimelineYearEntry | null>(
    null,
  );

  if (!years || years.length === 0) {
    return (
      <div className="flex h-36 items-center justify-center text-xs text-slate-500 border border-dashed border-[#1e293b] rounded">
        No dated release records found in catalog.
      </div>
    );
  }

  const maxGames = Math.max(...years.map((y) => y.game_count), 1);

  return (
    <div className="space-y-4">
      <div className="h-44 w-full flex items-end space-x-1.5 pt-6 pb-2 px-1 border-b border-[#1e293b]">
        {years.map((y) => {
          const heightPct = (y.game_count / maxGames) * 100;
          const isHovered = hoveredYear?.year === y.year;

          return (
            <div
              key={y.year}
              className="flex-1 flex flex-col items-center h-full justify-end group cursor-pointer"
              onMouseEnter={() => setHoveredYear(y)}
              onMouseLeave={() => setHoveredYear(null)}
            >
              <div
                className={`w-full max-w-[28px] rounded-t transition-all duration-200 ${
                  isHovered
                    ? "bg-cyan-400"
                    : "bg-cyan-500/70 group-hover:bg-cyan-400"
                }`}
                style={{ height: `${Math.max(4, heightPct)}%` }}
              />
              <span className="text-[10px] font-mono text-slate-500 mt-2 rotate-45 sm:rotate-0 origin-left sm:origin-center">
                {String(y.year).slice(-2)}
              </span>
            </div>
          );
        })}
      </div>

      {/* Selected / Hovered year detail banner */}
      <div className="h-10 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] flex items-center justify-between text-xs">
        {hoveredYear ? (
          <>
            <div className="flex items-center space-x-3">
              <span className="font-mono font-bold text-white text-sm">
                {hoveredYear.year}
              </span>
              <span className="text-slate-400">
                <strong className="text-white font-mono">
                  {hoveredYear.game_count}
                </strong>{" "}
                {hoveredYear.game_count === 1
                  ? "game released"
                  : "games released"}
              </span>
              <span className="text-slate-400">
                ({hoveredYear.paid_count} paid, {hoveredYear.free_count} free)
              </span>
            </div>
            <div className="font-mono text-cyan-400">
              {hoveredYear.owners_midpoint.toLocaleString()} est. owners
              midpoint
            </div>
          </>
        ) : (
          <span className="text-slate-400 text-xs italic">
            Hover over any year column to inspect releases and ownership totals.
          </span>
        )}
      </div>
    </div>
  );
};
