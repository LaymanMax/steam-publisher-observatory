import React from "react";
import type { PriceTierItem } from "../../types";

interface PriceTierStackedBarProps {
  tiers: PriceTierItem[];
}

export const PriceTierStackedBar: React.FC<PriceTierStackedBarProps> = ({
  tiers,
}) => {
  if (!tiers || tiers.length === 0) {
    return null;
  }

  const tierColors: Record<string, { bg: string; text: string }> = {
    Free: { bg: "bg-emerald-500", text: "text-emerald-400" },
    "Very Low": { bg: "bg-teal-500", text: "text-teal-400" },
    Low: { bg: "bg-cyan-500", text: "text-cyan-400" },
    "Mid-range": { bg: "bg-blue-500", text: "text-blue-400" },
    Premium: { bg: "bg-indigo-500", text: "text-indigo-400" },
    "High-priced": { bg: "bg-purple-500", text: "text-purple-400" },
    Missing: { bg: "bg-slate-600", text: "text-slate-400" },
  };

  return (
    <div className="space-y-4">
      {/* Segmented bar */}
      <div className="h-4 w-full bg-[#162032] rounded-md overflow-hidden flex">
        {tiers.map((t) => {
          if (t.share <= 0) return null;
          const color = tierColors[t.tier_name] || {
            bg: "bg-slate-500",
            text: "text-slate-400",
          };

          return (
            <div
              key={t.tier_name}
              className={`h-full ${color.bg} transition-all duration-300 relative group cursor-pointer`}
              style={{ width: `${t.share}%` }}
              title={`${t.tier_name}: ${t.game_count} games (${t.share.toFixed(1)}%)`}
            />
          );
        })}
      </div>

      {/* Legend / Breakdown table */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2 pt-1">
        {tiers.map((t) => {
          const color = tierColors[t.tier_name] || {
            bg: "bg-slate-500",
            text: "text-slate-400",
          };
          let rangeLabel = `$${t.min_price.toFixed(2)} - $${t.max_price.toFixed(2)}`;
          if (t.tier_name === "Free") rangeLabel = "$0.00";
          if (t.tier_name === "High-priced") rangeLabel = "$50.00+";
          if (t.tier_name === "Missing") rangeLabel = "Not recorded";

          return (
            <div
              key={t.tier_name}
              className="p-2 rounded border border-[#1e293b] bg-[#0d131f] flex flex-col justify-between"
            >
              <div className="flex items-center space-x-1.5 text-xs mb-1">
                <span className={`h-2 w-2 rounded-full ${color.bg} shrink-0`} />
                <span className="font-medium text-slate-300 truncate">
                  {t.tier_name}
                </span>
              </div>
              <div className="flex items-baseline justify-between text-xs font-mono">
                <span className="text-[11px] text-slate-500">{rangeLabel}</span>
                <span className="font-semibold text-white">
                  {t.share.toFixed(1)}%{" "}
                  <span className="text-[10px] text-slate-500 font-normal">
                    ({t.game_count})
                  </span>
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
