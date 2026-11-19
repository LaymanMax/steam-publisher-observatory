package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
)

// JobStatus represents state of an import job.
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

// ImportJob models a background import or download task.
type ImportJob struct {
	ID          string                `json:"id"`
	Type        string                `json:"type"` // "download" or "upload" or "reimport"
	Stage       importer.ImportStage  `json:"stage"`
	Progress    float64               `json:"progress"`
	Message     string                `json:"message"`
	Status      JobStatus             `json:"status"`
	Error       string                `json:"error,omitempty"`
	Stats       *importer.ImportStats `json:"stats,omitempty"`
	StartedAt   time.Time             `json:"started_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`

	cancelFunc context.CancelFunc `json:"-"`
}

// Manager coordinates running and completed import jobs.
type Manager struct {
	importer *importer.Importer
	jobs     map[string]*ImportJob
	activeID string
	mu       sync.RWMutex
}

// NewManager creates a Job Manager.
func NewManager(imp *importer.Importer) *Manager {
	return &Manager{
		importer: imp,
		jobs:     make(map[string]*ImportJob),
	}
}

// CreateJob registers a new pending job.
func (m *Manager) CreateJob(jobType string) (*ImportJob, context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	job := &ImportJob{
		ID:         uuid.New().String(),
		Type:       jobType,
		Stage:      importer.StageValidating,
		Progress:   0.0,
		Message:    "Job initialized...",
		Status:     StatusPending,
		StartedAt:  time.Now(),
		cancelFunc: cancel,
	}

	m.jobs[job.ID] = job
	m.activeID = job.ID
	return job, ctx
}

// GetJob returns a job by ID.
func (m *Manager) GetJob(id string) (*ImportJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[id]
	return j, ok
}

// GetActiveJob returns the currently active or most recent job.
func (m *Manager) GetActiveJob() *ImportJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activeID == "" {
		return nil
	}
	return m.jobs[m.activeID]
}

// UpdateProgress updates stage, percentage, and message for a job.
func (m *Manager) UpdateProgress(jobID string, stage importer.ImportStage, progress float64, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[jobID]; ok {
		job.Status = StatusRunning
		job.Stage = stage
		job.Progress = progress
		job.Message = message
	}
}

// MarkCompleted sets status to completed with statistics.
func (m *Manager) MarkCompleted(jobID string, stats *importer.ImportStats) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[jobID]; ok {
		now := time.Now()
		job.Status = StatusCompleted
		job.Stage = importer.StageComplete
		job.Progress = 100.0
		job.Message = "Dataset import completed successfully."
		job.Stats = stats
		job.CompletedAt = &now
	}
}

// MarkFailed records a failure.
func (m *Manager) MarkFailed(jobID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[jobID]; ok {
		now := time.Now()
		job.Status = StatusFailed
		job.Stage = importer.StageFailed
		job.Error = err.Error()
		job.Message = fmt.Sprintf("Import failed: %v", err)
		job.CompletedAt = &now
	}
}

// Cancel cancels a job if running.
func (m *Manager) Cancel(jobID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[jobID]; ok && job.Status == StatusRunning {
		if job.cancelFunc != nil {
			job.cancelFunc()
		}
		now := time.Now()
		job.Status = StatusFailed
		job.Error = "Job cancelled by user"
		job.Message = "Job cancelled"
		job.CompletedAt = &now
		return true
	}
	return false
}
