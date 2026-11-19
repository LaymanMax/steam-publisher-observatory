package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
)

func (s *Server) handleGetOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := s.analytics.GetOverview(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, overview)
}

func (s *Server) handleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	overview, err := s.analytics.GetOverview(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	activeJob := s.jobs.GetActiveJob()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_games":        overview.TotalGames,
		"total_publishers":   overview.TotalPublishers,
		"total_developers":   overview.TotalDevelopers,
		"ownership_coverage": overview.OwnershipCoverage,
		"dataset_freshness":  overview.DatasetFreshness,
		"source":             overview.Source,
		"active_job":         activeJob,
	})
}

func (s *Server) handleDatasetImports(w http.ResponseWriter, r *http.Request) {
	// List imports
	activeJob := s.jobs.GetActiveJob()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"active_job": activeJob,
	})
}

func (s *Server) handleDatasetImport(w http.ResponseWriter, r *http.Request) {
	if job := s.jobs.GetActiveJob(); job != nil && job.Status == "running" {
		respondError(w, http.StatusConflict, "An import job is already running")
		return
	}

	job, jobCtx := s.jobs.CreateJob("download")

	go func() {
		// Check if local parquet already exists in data directory
		localParquet := filepath.Join(s.config.DataDir, "train-00000-of-00001.parquet")
		if _, err := os.Stat(localParquet); err == nil {
			s.jobs.UpdateProgress(job.ID, importer.StageValidating, 10.0, "Found local Parquet file. Validating schema...")
			stats, err := s.importer.Ingest(jobCtx, localParquet, "local:train-00000-of-00001.parquet", func(stage importer.ImportStage, pct float64, msg string) {
				s.jobs.UpdateProgress(job.ID, stage, pct, msg)
			})
			if err != nil {
				s.jobs.MarkFailed(job.ID, err)
			} else {
				s.jobs.MarkCompleted(job.ID, stats)
			}
			return
		}

		// Otherwise, discover and download from Hugging Face
		s.jobs.UpdateProgress(job.ID, importer.StageValidating, 5.0, "Discovering Parquet files on Hugging Face...")
		discovered, err := s.hfClient.DiscoverParquetFiles(jobCtx)
		if err != nil || len(discovered) == 0 {
			s.jobs.MarkFailed(job.ID, fmt.Errorf("failed to discover dataset files on Hugging Face: %w", err))
			return
		}

		targetFile := discovered[0]
		s.jobs.UpdateProgress(job.ID, importer.StageValidating, 10.0, fmt.Sprintf("Downloading %s (%.1f MB)...", targetFile.Path, float64(targetFile.Size)/(1024*1024)))

		destPath := filepath.Join(s.config.DataDir, filepath.Base(targetFile.Path))
		_, err = s.hfClient.DownloadFile(jobCtx, targetFile.DownloadURL, destPath, func(downloaded, total int64, pct float64) {
			s.jobs.UpdateProgress(job.ID, importer.StageValidating, 10.0+(pct*0.2), fmt.Sprintf("Downloading dataset: %.1f%%", pct))
		})
		if err != nil {
			s.jobs.MarkFailed(job.ID, fmt.Errorf("download failed: %w", err))
			return
		}

		// Ingest downloaded file
		stats, err := s.importer.Ingest(jobCtx, destPath, "huggingface:"+targetFile.Path, func(stage importer.ImportStage, pct float64, msg string) {
			s.jobs.UpdateProgress(job.ID, stage, pct, msg)
		})
		if err != nil {
			s.jobs.MarkFailed(job.ID, err)
		} else {
			s.jobs.MarkCompleted(job.ID, stats)
		}
	}()

	respondJSON(w, http.StatusAccepted, job)
}

func (s *Server) handleDatasetUpload(w http.ResponseWriter, r *http.Request) {
	if job := s.jobs.GetActiveJob(); job != nil && job.Status == "running" {
		respondError(w, http.StatusConflict, "An import job is already running")
		return
	}

	// 500 MB max upload
	if err := r.ParseMultipartForm(500 * 1024 * 1024); err != nil {
		respondError(w, http.StatusBadRequest, "File size exceeds limit or malformed multipart request")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided in form field 'file'")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".parquet") {
		respondError(w, http.StatusBadRequest, "Only .parquet files are supported")
		return
	}

	job, jobCtx := s.jobs.CreateJob("upload")
	destPath := filepath.Join(s.config.DataDir, "upload_"+header.Filename)

	go func() {
		s.jobs.UpdateProgress(job.ID, importer.StageValidating, 5.0, "Saving uploaded Parquet file...")
		out, err := os.Create(destPath)
		if err != nil {
			s.jobs.MarkFailed(job.ID, fmt.Errorf("failed to save upload: %w", err))
			return
		}

		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			s.jobs.MarkFailed(job.ID, fmt.Errorf("upload transfer failed: %w", err))
			return
		}
		out.Close()

		stats, err := s.importer.Ingest(jobCtx, destPath, "upload:"+header.Filename, func(stage importer.ImportStage, pct float64, msg string) {
			s.jobs.UpdateProgress(job.ID, stage, pct, msg)
		})
		if err != nil {
			s.jobs.MarkFailed(job.ID, err)
		} else {
			s.jobs.MarkCompleted(job.ID, stats)
		}
	}()

	respondJSON(w, http.StatusAccepted, job)
}

func (s *Server) handleDatasetCancel(w http.ResponseWriter, r *http.Request) {
	if job := s.jobs.GetActiveJob(); job != nil {
		cancelled := s.jobs.Cancel(job.ID)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"cancelled": cancelled,
			"job_id":    job.ID,
		})
		return
	}
	respondError(w, http.StatusNotFound, "No active job found to cancel")
}

func (s *Server) handleDatasetSchema(w http.ResponseWriter, r *http.Request) {
	// Inspect current parquet in data directory
	localParquet := filepath.Join(s.config.DataDir, "train-00000-of-00001.parquet")
	count, cols, err := s.importer.InspectParquet(r.Context(), localParquet)
	if err != nil {
		respondError(w, http.StatusNotFound, "No local Parquet dataset found to inspect")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"row_count": count,
		"columns":   cols,
	})
}

func (s *Server) handleDatasetQuality(w http.ResponseWriter, r *http.Request) {
	diag, err := s.analytics.GetDataQualityDiagnostics(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, diag)
}

func (s *Server) handleDatasetIssues(w http.ResponseWriter, r *http.Request) {
	issue := r.URL.Query().Get("issue")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	records, total, err := s.analytics.ListProblematicRecords(r.Context(), issue, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
