import React, { useState } from "react";
import type { GenreItem } from "../../types";

interface HorizontalBarChartProps {
  items: GenreItem[];
  maxDisplay?: number;
}

export const HorizontalBarChart: React.FC<HorizontalBarChartProps> = ({
  items,
  maxDisplay = 10,
}) => {
  const [hoveredGenre, setHoveredGenre] = useState<string | null>(null);

  if (!items || items.length === 0) {
    return (
      <div className="flex h-36 items-center justify-center text-xs text-slate-500 border border-dashed border-[#1e293b] rounded">
        No genre information available.
      </div>
    );
  }

  const displayed = items.slice(0, maxDisplay);

  return (
    <div className="space-y-2.5">
      {displayed.map((g) => {
        const isHovered = hoveredGenre === g.genre;

        return (
          <div
            key={g.genre}
            className="group cursor-pointer"
            onMouseEnter={() => setHoveredGenre(g.genre)}
            onMouseLeave={() => setHoveredGenre(null)}
          >
            <div className="flex items-center justify-between text-xs mb-1">
              <span
                className={`font-medium transition-standard ${isHovered ? "text-cyan-400" : "text-slate-300"}`}
              >
                {g.genre}
              </span>
              <div className="flex items-center space-x-3 text-[11px] font-mono">
                <span className="text-slate-500">
                  {g.game_count} {g.game_count === 1 ? "title" : "titles"}
                </span>
                <span className="font-semibold text-slate-200">
                  {g.weighted_share.toFixed(1)}%
                </span>
              </div>
            </div>

            <div className="h-2 w-full bg-[#162032] rounded-sm overflow-hidden">
              <div
                className={`h-full rounded-sm transition-all duration-300 ${
                  isHovered
                    ? "bg-cyan-400"
                    : "bg-cyan-500/80 group-hover:bg-cyan-400"
                }`}
                style={{
                  width: `${Math.min(100, Math.max(1, g.weighted_share))}%`,
                }}
              />
            </div>
          </div>
        );
      })}

      {items.length > maxDisplay && (
        <div className="text-center text-xs text-slate-500 pt-1">
          + {items.length - maxDisplay} additional genres
        </div>
      )}
    </div>
  );
};
