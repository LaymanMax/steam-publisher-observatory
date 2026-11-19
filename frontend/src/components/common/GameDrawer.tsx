import {
  Activity,
  Calendar,
  DollarSign,
  ExternalLink,
  ShieldCheck,
  ThumbsUp,
  Users,
  X,
} from "lucide-react";
import React from "react";
import type { GameCatalogItem } from "../../types";

interface GameDrawerProps {
  game: GameCatalogItem | null;
  onClose: () => void;
}

export const GameDrawer: React.FC<GameDrawerProps> = ({ game, onClose }) => {
  if (!game) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-black/60 backdrop-blur-sm flex justify-end">
      <div className="w-full max-w-lg bg-[#0d131f] border-l border-[#1e293b] h-full overflow-y-auto p-6 shadow-2xl flex flex-col justify-between">
        <div className="space-y-6">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div>
              <span className="font-mono text-xs text-slate-500">
                AppID: {game.appid}
              </span>
              <h2 className="text-lg font-bold text-white mt-1 leading-snug">
                {game.name}
              </h2>
            </div>
            <button
              onClick={onClose}
              className="p-1 rounded text-slate-400 hover:text-white hover:bg-[#162032] transition-standard"
            >
              <X className="h-5 w-5" />
            </button>
          </div>

          {/* Header image with fallback */}
          <div className="relative aspect-video w-full rounded-md overflow-hidden bg-[#162032] border border-[#1e293b]">
            {game.header_image ? (
              <img
                src={game.header_image}
                alt={game.name}
                className="w-full h-full object-cover"
                onError={(e) => {
                  (e.target as HTMLElement).style.display = "none";
                }}
              />
            ) : (
              <div className="flex h-full items-center justify-center text-xs text-slate-500">
                No promotional image provided in dataset
              </div>
            )}
          </div>

          {/* Short description */}
          {game.short_description && (
            <p className="text-xs text-slate-300 leading-relaxed bg-[#111827] p-3 rounded border border-[#1e293b]">
              {game.short_description}
            </p>
          )}

          {/* Key Metrics Grid */}
          <div className="grid grid-cols-2 gap-3 text-xs">
            {/* Release Date */}
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="flex items-center text-slate-400 text-[11px]">
                <Calendar className="h-3 w-3 mr-1.5 text-cyan-400" />
                Release Date
              </span>
              <span className="font-mono font-semibold text-white">
                {game.release_date || "Unknown"}
              </span>
            </div>

            {/* Price */}
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="flex items-center text-slate-400 text-[11px]">
                <DollarSign className="h-3 w-3 mr-1.5 text-emerald-400" />
                Price ({game.price_tier})
              </span>
              <span className="font-mono font-semibold text-white">
                {game.price !== undefined
                  ? `$${game.price.toFixed(2)}`
                  : "Missing"}
              </span>
            </div>

            {/* Ownership Range */}
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1 col-span-2">
              <span className="flex items-center text-slate-400 text-[11px]">
                <Users className="h-3 w-3 mr-1.5 text-blue-400" />
                Estimated Owners Range
              </span>
              <div className="flex items-baseline justify-between font-mono">
                <span className="text-white font-medium">
                  {game.owners_raw || "Missing"}
                </span>
                {game.owners_midpoint !== undefined && (
                  <span className="text-cyan-400 font-bold">
                    ~{game.owners_midpoint.toLocaleString()} midpoint
                  </span>
                )}
              </div>
            </div>

            {/* Reviews */}
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="flex items-center text-slate-400 text-[11px]">
                <ThumbsUp className="h-3 w-3 mr-1.5 text-emerald-400" />
                Reviews (
                {game.positive_share
                  ? `${game.positive_share.toFixed(1)}% pos.`
                  : "N/A"}
                )
              </span>
              <div className="font-mono text-white text-[11px]">
                <span className="text-emerald-400 font-semibold">
                  {game.positive.toLocaleString()}
                </span>{" "}
                /{" "}
                <span className="text-rose-400">
                  {game.negative.toLocaleString()}
                </span>
              </div>
            </div>

            {/* CCU & Playtime */}
            <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-1">
              <span className="flex items-center text-slate-400 text-[11px]">
                <Activity className="h-3 w-3 mr-1.5 text-amber-400" />
                Peak CCU / Playtime
              </span>
              <div className="font-mono text-white text-[11px]">
                {game.peak_ccu.toLocaleString()} ccu |{" "}
                {Math.round(game.median_playtime / 60)}h med.
              </div>
            </div>
          </div>

          {/* Metadata Quality */}
          <div className="p-3 rounded bg-[#111827] border border-[#1e293b] space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="flex items-center text-slate-400 text-[11px]">
                <ShieldCheck className="h-3.5 w-3.5 mr-1.5 text-cyan-400" />
                Metadata Completeness Score
              </span>
              <span className="font-mono font-bold text-white">
                {game.quality_score.toFixed(0)}%
              </span>
            </div>
            <div className="h-1.5 w-full bg-[#1e293b] rounded-full overflow-hidden">
              <div
                className="h-full bg-cyan-400 rounded-full"
                style={{ width: `${game.quality_score}%` }}
              />
            </div>
            {game.is_playtest && (
              <span className="inline-block px-2 py-0.5 rounded text-[10px] font-mono bg-rose-500/10 text-rose-300 border border-rose-500/20">
                Suspected Playtest / Non-Commercial
              </span>
            )}
          </div>
        </div>

        {/* Footer actions */}
        <div className="pt-6 border-t border-[#1e293b] flex items-center justify-between">
          <a
            href={`https://store.steampowered.com/app/${game.appid}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded bg-[#162032] border border-[#23334d] text-cyan-400 hover:bg-[#1e293b] transition-standard text-xs font-medium"
          >
            <span>View on Steam Store</span>
            <ExternalLink className="h-3.5 w-3.5" />
          </a>
          <button
            onClick={onClose}
            className="px-4 py-1.5 rounded text-xs font-medium text-slate-400 hover:text-white"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
