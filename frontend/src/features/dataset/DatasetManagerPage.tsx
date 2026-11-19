import {
  AlertCircle,
  CheckCircle2,
  Database,
  Download,
  Layers,
  RefreshCw,
  Upload,
  XCircle,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../services/api";
import type { ImportJob } from "../../types";

export const DatasetManagerPage: React.FC = () => {
  const [status, setStatus] = useState<{
    total_games: number;
    total_publishers: number;
    total_developers: number;
    ownership_coverage: number;
    dataset_freshness?: string;
    source?: string;
    active_job?: ImportJob;
  } | null>(null);

  const [schemaCols, setSchemaCols] = useState<string[]>([]);
  const [schemaRowCount, setSchemaRowCount] = useState<number>(0);
  const [activeJob, setActiveJob] = useState<ImportJob | null>(null);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  useEffect(() => {
    loadStatus();
    loadSchema();
  }, []);

  // Poll for active job progress if running
  useEffect(() => {
    let interval: any = null;
    if (
      activeJob &&
      (activeJob.status === "running" || activeJob.status === "pending")
    ) {
      interval = setInterval(async () => {
        try {
          const res = await api.getDatasetStatus();
          setStatus(res);
          if (res.active_job) {
            setActiveJob(res.active_job);
            if (
              res.active_job.status === "completed" ||
              res.active_job.status === "failed"
            ) {
              loadSchema();
            }
          }
        } catch {
          // ignore
        }
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [activeJob?.status]);

  const loadStatus = async () => {
    try {
      const res = await api.getDatasetStatus();
      setStatus(res);
      if (res.active_job) {
        setActiveJob(res.active_job);
      }
    } catch (err) {
      console.error(err);
    }
  };

  const loadSchema = async () => {
    try {
      const s = await api.getDatasetSchema();
      setSchemaCols(s.columns || []);
      setSchemaRowCount(s.row_count || 0);
    } catch {
      // not yet available
    }
  };

  const handleStartImport = async () => {
    try {
      setIsSubmitting(true);
      setUploadError(null);
      const job = await api.startImport();
      setActiveJob(job);
    } catch (err: any) {
      setUploadError(err.message || "Failed to start import job");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.name.toLowerCase().endsWith(".parquet")) {
      setUploadError("Only .parquet dataset files are accepted.");
      return;
    }

    try {
      setIsSubmitting(true);
      setUploadError(null);
      const job = await api.uploadParquet(file);
      setActiveJob(job);
    } catch (err: any) {
      setUploadError(err.message || "Failed to upload Parquet file");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancelJob = async () => {
    try {
      await api.cancelImport();
      loadStatus();
    } catch (err) {
      console.error(err);
    }
  };

  const isJobActive =
    activeJob &&
    (activeJob.status === "running" || activeJob.status === "pending");

  return (
    <div className="container-max py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-[#1e293b] pb-6">
        <h1 className="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
          <Database className="h-5 w-5 text-cyan-400" />
          <span>Dataset Acquisition & Ingestion Manager</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Manage Parquet dataset downloads, local file imports, and inspect
          schema discovery.
        </p>
      </div>

      {/* Active Job Progress Tracker */}
      {activeJob && (
        <div className="rounded-xl border border-cyan-500/30 bg-[#111928] p-6 space-y-4 shadow-lg">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              {activeJob.status === "completed" ? (
                <CheckCircle2 className="h-5 w-5 text-emerald-400" />
              ) : activeJob.status === "failed" ? (
                <AlertCircle className="h-5 w-5 text-rose-400" />
              ) : (
                <RefreshCw className="h-5 w-5 text-cyan-400 animate-spin" />
              )}
              <span className="font-bold text-white text-sm">
                Import Job:{" "}
                <span className="uppercase font-mono text-cyan-400">
                  {activeJob.stage}
                </span>
              </span>
            </div>

            <div className="flex items-center space-x-3">
              <span className="font-mono text-xs font-semibold text-white">
                {activeJob.progress.toFixed(0)}%
              </span>
              {isJobActive && (
                <button
                  onClick={handleCancelJob}
                  className="inline-flex items-center space-x-1 px-2.5 py-1 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20 text-xs hover:bg-rose-500/20"
                >
                  <XCircle className="h-3.5 w-3.5" />
                  <span>Cancel Job</span>
                </button>
              )}
            </div>
          </div>

          {/* Progress Bar */}
          <div className="h-2 w-full bg-[#1e293b] rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full transition-all duration-300 ${
                activeJob.status === "completed"
                  ? "bg-emerald-400"
                  : activeJob.status === "failed"
                    ? "bg-rose-500"
                    : "bg-gradient-to-r from-cyan-500 to-blue-500"
              }`}
              style={{ width: `${Math.max(2, activeJob.progress)}%` }}
            />
          </div>

          <div className="flex items-center justify-between text-xs text-slate-400 font-mono">
            <span>{activeJob.message}</span>
            <span>
              Started: {new Date(activeJob.started_at).toLocaleTimeString()}
            </span>
          </div>

          {activeJob.error && (
            <div className="p-3 rounded bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs">
              <strong>Error:</strong> {activeJob.error}
            </div>
          )}
        </div>
      )}

      {/* Dataset Actions Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Option A: Download from Hugging Face */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center space-x-2 pb-2 border-b border-[#1e293b]">
            <Download className="h-4 w-4 text-cyan-400" />
            <h2 className="text-sm font-bold text-white uppercase font-mono">
              Option A: Hugging Face Discovery & Download
            </h2>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed">
            Discovers the latest Parquet dataset file from repository{" "}
            <code>FronkonGames/steam-games-dataset</code> using the official
            Hugging Face Hub metadata API, validates SHA-256 integrity, and
            imports directly into embedded DuckDB storage.
          </p>
          <div className="pt-2">
            <button
              onClick={handleStartImport}
              disabled={isJobActive || isSubmitting}
              className="w-full sm:w-auto inline-flex items-center justify-center space-x-2 px-4 py-2 rounded bg-cyan-500 text-slate-950 font-semibold text-xs hover:bg-cyan-400 disabled:opacity-50 disabled:cursor-not-allowed transition-standard"
            >
              <RefreshCw
                className={`h-3.5 w-3.5 ${isSubmitting ? "animate-spin" : ""}`}
              />
              <span>
                {isJobActive
                  ? "Import in Progress..."
                  : "Start Dataset Ingestion"}
              </span>
            </button>
          </div>
        </div>

        {/* Option B: Upload Local Parquet */}
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center space-x-2 pb-2 border-b border-[#1e293b]">
            <Upload className="h-4 w-4 text-emerald-400" />
            <h2 className="text-sm font-bold text-white uppercase font-mono">
              Option B: Upload Local Parquet File
            </h2>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed">
            Provide your own copy of the Steam games dataset in Apache Parquet
            format. The backend validates schema conformity and normalizes
            entities automatically.
          </p>
          <div className="pt-2">
            <label className="w-full sm:w-auto inline-flex items-center justify-center space-x-2 px-4 py-2 rounded bg-[#162032] border border-[#23334d] text-slate-200 font-semibold text-xs hover:bg-[#1f2d47] cursor-pointer transition-standard">
              <Upload className="h-3.5 w-3.5 text-emerald-400" />
              <span>Choose .parquet file to upload</span>
              <input
                type="file"
                accept=".parquet"
                onChange={handleFileUpload}
                disabled={isJobActive || isSubmitting}
                className="hidden"
              />
            </label>
          </div>
        </div>
      </div>

      {uploadError && (
        <div className="p-3 rounded bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs">
          {uploadError}
        </div>
      )}

      {/* Discovered Parquet Schema Inspector */}
      {schemaCols.length > 0 && (
        <div className="rounded-lg border border-[#1e293b] bg-[#111827] p-5 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-[#1e293b]">
            <div>
              <h2 className="text-xs font-semibold text-white uppercase font-mono tracking-wide flex items-center">
                <Layers className="h-4 w-4 mr-1.5 text-cyan-400" />
                Discovered Parquet Schema
              </h2>
              <p className="text-[11px] text-slate-400">
                Inspected from verified dataset file (
                {schemaRowCount.toLocaleString()} records).
              </p>
            </div>
            <span className="text-xs font-mono text-cyan-400">
              {schemaCols.length} columns detected
            </span>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2">
            {schemaCols.map((col) => (
              <div
                key={col}
                className="p-2 rounded bg-[#090d16] border border-[#1e293b] text-xs font-mono text-slate-300 truncate"
                title={col}
              >
                {col}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
