package service

import (
	"sync"
	"time"

	"ai-service-go/internal/dto"
)

type jobStateStore struct {
	mu   sync.RWMutex
	jobs map[string]*dto.JobDebugResponse
}

func newJobStateStore() *jobStateStore {
	return &jobStateStore{
		jobs: map[string]*dto.JobDebugResponse{},
	}
}

func (s *jobStateStore) initJob(jobID, jobType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[jobID] = &dto.JobDebugResponse{
		JobID:          jobID,
		JobType:        jobType,
		Status:         "queued",
		CallbackStatus: "pending",
		UpdatedAt:      time.Now(),
	}
}

func (s *jobStateStore) update(jobID string, mutate func(job *dto.JobDebugResponse)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		job = &dto.JobDebugResponse{
			JobID:          jobID,
			Status:         "unknown",
			CallbackStatus: "pending",
		}
		s.jobs[jobID] = job
	}

	mutate(job)
	job.UpdatedAt = time.Now()
}

func (s *jobStateStore) get(jobID string) (*dto.JobDebugResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil, false
	}

	cloned := *job
	cloned.SearchAttempts = append([]dto.SearchAttempt(nil), job.SearchAttempts...)
	cloned.TinyfishFetches = append([]string(nil), job.TinyfishFetches...)
	return &cloned, true
}
