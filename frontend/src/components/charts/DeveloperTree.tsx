import {
  ChevronDown,
  ChevronRight,
  Gamepad2,
  Share2,
  Users,
} from "lucide-react";
import React, { useState } from "react";
import type { DeveloperAnalysisReport } from "../../types";

interface DeveloperTreeProps {
  report: DeveloperAnalysisReport;
}

export const DeveloperTree: React.FC<DeveloperTreeProps> = ({ report }) => {
  const [expandedDevs, setExpandedDevs] = useState<Record<string, boolean>>({});

  const toggleDev = (devId: string) => {
    setExpandedDevs((prev) => ({ ...prev, [devId]: !prev[devId] }));
  };

  if (!report || report.developers.length === 0) {
    return (
      <div className="flex h-36 items-center justify-center text-xs text-slate-500 border border-dashed border-[#1e293b] rounded">
        No developer relationships recorded for this publisher.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Overview Badges */}
      <div className="flex flex-wrap items-center gap-3 text-xs">
        <span className="flex items-center space-x-1.5 px-2.5 py-1 rounded bg-[#162032] border border-[#23334d] text-slate-300">
          <Users className="h-3.5 w-3.5 text-cyan-400" />
          <span>
            <strong>{report.total_developers}</strong> distinct studios
          </span>
        </span>
        <span className="px-2.5 py-1 rounded bg-[#162032] border border-[#23334d] text-slate-300">
          <strong>{report.multi_title_developers}</strong> recurring studios
        </span>
        <span className="px-2.5 py-1 rounded bg-[#162032] border border-[#23334d] text-slate-300">
          <strong>{report.single_title_developers}</strong> single-title studios
        </span>
        {report.top_developer_name && (
          <span className="px-2.5 py-1 rounded bg-[#162032] border border-[#23334d] text-slate-300">
            Primary studio:{" "}
            <strong className="text-cyan-400">
              {report.top_developer_name}
            </strong>{" "}
            ({report.top_developer_share.toFixed(1)}% of catalog)
          </span>
        )}
      </div>

      {/* Developer Hierarchy List */}
      <div className="space-y-2 max-h-[500px] overflow-y-auto pr-1">
        {report.developers.map((dev) => {
          const isExpanded = !!expandedDevs[dev.developer_id];

          return (
            <div
              key={dev.developer_id}
              className="rounded-md border border-[#1e293b] bg-[#0d131f] overflow-hidden transition-standard"
            >
              <div
                onClick={() => toggleDev(dev.developer_id)}
                className="flex items-center justify-between p-3 cursor-pointer hover:bg-[#162032] transition-standard"
              >
                <div className="flex items-center space-x-2.5">
                  {isExpanded ? (
                    <ChevronDown className="h-4 w-4 text-slate-400 shrink-0" />
                  ) : (
                    <ChevronRight className="h-4 w-4 text-slate-400 shrink-0" />
                  )}
                  <span className="font-semibold text-white text-xs">
                    {dev.developer_name}
                  </span>
                  {dev.shared_publishers > 0 && (
                    <span
                      className="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded text-[10px] font-mono bg-purple-500/10 text-purple-300 border border-purple-500/20"
                      title={`Also developed games published by ${dev.shared_publishers} other publisher(s)`}
                    >
                      <Share2 className="h-2.5 w-2.5" />
                      <span>+{dev.shared_publishers} shared</span>
                    </span>
                  )}
                </div>

                <div className="flex items-center space-x-4 text-xs font-mono">
                  <span className="text-slate-400">
                    <strong className="text-white">
                      {dev.associated_games}
                    </strong>{" "}
                    {dev.associated_games === 1 ? "title" : "titles"} (
                    {dev.share_of_catalog.toFixed(1)}%)
                  </span>
                  <span className="text-cyan-400 font-semibold">
                    {dev.owners_midpoint.toLocaleString()} est. midpoint
                  </span>
                </div>
              </div>

              {/* Expanded games list */}
              {isExpanded && (
                <div className="bg-[#090d16] p-3 border-t border-[#1e293b] space-y-1.5">
                  <div className="text-[11px] font-mono uppercase text-slate-500 mb-2">
                    Associated Titles ({dev.games?.length || 0})
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                    {dev.games?.map((g) => (
                      <div
                        key={g.appid}
                        className="p-2 rounded bg-[#111827] border border-[#1e293b] flex items-center justify-between text-xs"
                      >
                        <div className="flex items-center space-x-2 truncate">
                          <Gamepad2 className="h-3.5 w-3.5 text-slate-500 shrink-0" />
                          <span
                            className="truncate text-slate-200"
                            title={g.name}
                          >
                            {g.name}
                          </span>
                        </div>
                        <div className="flex items-center space-x-2 font-mono text-[11px] shrink-0 text-slate-400">
                          {g.release_year && <span>{g.release_year}</span>}
                          <span className="text-cyan-400 font-medium">
                            {g.midpoint.toLocaleString()} est.
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>

      <div className="text-[11px] text-slate-500 italic pt-1">
        * {report.notice}
      </div>
    </div>
  );
};
