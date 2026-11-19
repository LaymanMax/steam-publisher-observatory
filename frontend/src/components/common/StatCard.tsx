import { HelpCircle } from "lucide-react";
import React from "react";

interface StatCardProps {
  label: string;
  value: string | number;
  subtext?: string;
  tooltip?: string;
  badge?: {
    text: string;
    variant: "neutral" | "success" | "warning" | "danger" | "info";
  };
}

export const StatCard: React.FC<StatCardProps> = ({
  label,
  value,
  subtext,
  tooltip,
  badge,
}) => {
  const badgeClasses = {
    neutral: "bg-slate-800 text-slate-300 border-slate-700",
    success: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    warning: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    danger: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    info: "bg-cyan-500/10 text-cyan-400 border-cyan-500/20",
  }[badge?.variant || "neutral"];

  return (
    <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-4 shadow-sm hover:border-[#2b3a55] transition-standard">
      <div className="flex items-center justify-between text-xs text-slate-400 mb-1.5">
        <span className="font-medium tracking-wide uppercase text-[11px] text-slate-400 flex items-center">
          {label}
          {tooltip && (
            <span
              className="ml-1 text-slate-500 hover:text-cyan-400 cursor-help"
              title={tooltip}
            >
              <HelpCircle className="h-3 w-3" />
            </span>
          )}
        </span>
        {badge && (
          <span
            className={`px-1.5 py-0.5 rounded text-[10px] font-mono border ${badgeClasses}`}
          >
            {badge.text}
          </span>
        )}
      </div>

      <div className="text-2xl font-bold font-mono tracking-tight text-white">
        {value}
      </div>

      {subtext && (
        <div className="mt-1 text-xs text-slate-400 font-normal">{subtext}</div>
      )}
    </div>
  );
};
