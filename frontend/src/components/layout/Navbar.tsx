import {
  BarChart3,
  Building2,
  Database,
  GitCompare,
  Layers,
  ShieldAlert,
} from "lucide-react";
import React from "react";
import { Link, useLocation } from "react-router-dom";

export const Navbar: React.FC = () => {
  const location = useLocation();

  const navItems = [
    { path: "/", label: "Observatory", icon: BarChart3 },
    { path: "/publishers", label: "Publishers Directory", icon: Building2 },
    { path: "/compare", label: "Compare", icon: GitCompare },
    { path: "/data-quality", label: "Data Quality", icon: ShieldAlert },
    { path: "/dataset", label: "Dataset Manager", icon: Database },
  ];

  return (
    <header className="sticky top-0 z-40 border-b border-[#1e293b] bg-[#090d16]/90 backdrop-blur-md">
      <div className="container-max flex h-14 items-center justify-between">
        <div className="flex items-center space-x-8">
          <Link
            to="/"
            className="flex items-center space-x-2.5 text-white no-underline"
          >
            <div className="flex h-7 w-7 items-center justify-center rounded bg-gradient-to-br from-cyan-500 to-blue-600 shadow-sm">
              <Layers className="h-4 w-4 text-white" />
            </div>
            <div>
              <span className="font-semibold tracking-tight text-white text-sm">
                Steam Publisher
              </span>
              <span className="ml-1.5 text-xs font-mono uppercase tracking-wider text-cyan-400 font-semibold">
                Observatory
              </span>
            </div>
          </Link>

          <nav className="hidden md:flex items-center space-x-1">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive =
                item.path === "/"
                  ? location.pathname === "/"
                  : location.pathname.startsWith(item.path);

              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`flex items-center space-x-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-standard ${
                    isActive
                      ? "bg-[#162032] text-cyan-400 border border-[#23334d]"
                      : "text-slate-400 hover:text-slate-200 hover:bg-[#111827]"
                  }`}
                >
                  <Icon className="h-3.5 w-3.5" />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="flex items-center space-x-3 text-xs">
          <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 mr-1.5 animate-pulse"></span>
            Live Parquet OLAP
          </span>
        </div>
      </div>
    </header>
  );
};
