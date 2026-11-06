package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/orchestrator"
)

// ScheduledJob represents a drill recurring on a cron pattern.
type ScheduledJob struct {
	DrillConfig config.DrillConfig `json:"drill_config"`
	Schedule    string             `json:"schedule"`
	LastRunAt   *time.Time         `json:"last_run_at,omitempty"`
	NextRunAt   time.Time          `json:"next_run_at"`
	Running     bool               `json:"running"`
}

// Scheduler handles background automated execution of scheduled disaster recovery drills.
type Scheduler struct {
	orch   *orchestrator.Orchestrator
	drills []config.DrillConfig
	jobs   map[string]*ScheduledJob
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewScheduler creates a scheduler for configured drill crons.
func NewScheduler(orch *orchestrator.Orchestrator, drills []config.DrillConfig) *Scheduler {
	s := &Scheduler{
		orch:   orch,
		drills: drills,
		jobs:   make(map[string]*ScheduledJob),
		stopCh: make(chan struct{}),
	}
	s.initJobs()
	return s
}

func (s *Scheduler) initJobs() {
	now := time.Now().UTC()
	for _, d := range s.drills {
		if d.Schedule != "" {
			nextRun, err := NextCronTime(d.Schedule, now)
			if err != nil {
				// Fallback to 24 hours if parsing fails
				nextRun = now.Add(24 * time.Hour)
			}
			s.jobs[d.ID] = &ScheduledJob{
				DrillConfig: d,
				Schedule:    d.Schedule,
				NextRunAt:   nextRun,
			}
		}
	}
}

// Start begins background polling for scheduled drills every 30 seconds.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-s.stopCh:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			case now := <-ticker.C:
				s.checkAndTrigger(ctx, now.UTC())
			}
		}
	}()
}

func (s *Scheduler) checkAndTrigger(ctx context.Context, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		if !job.Running && (now.After(job.NextRunAt) || now.Equal(job.NextRunAt)) {
			job.Running = true
			runTime := now
			job.LastRunAt = &runTime

			// Calculate next execution
			next, err := NextCronTime(job.Schedule, now)
			if err != nil {
				next = now.Add(24 * time.Hour)
			}
			job.NextRunAt = next

			// Execute asynchronously
			go func(j *ScheduledJob) {
				_, _ = s.orch.RunDrill(ctx, j.DrillConfig)
				s.mu.Lock()
				j.Running = false
				s.mu.Unlock()
			}(job)
		}
	}
}

// GetJobs returns status of all registered scheduled drills.
func (s *Scheduler) GetJobs() []ScheduledJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]ScheduledJob, 0, len(s.jobs))
	for _, j := range s.jobs {
		list = append(list, *j)
	}
	return list
}

// Stop cleanly terminates the scheduler ticker.
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// NextCronTime computes the next trigger time for standard 5-part cron expressions (minute hour dom month dow).
func NextCronTime(cronExpr string, from time.Time) (time.Time, error) {
	cronExpr = strings.TrimSpace(cronExpr)
	if cronExpr == "@daily" || cronExpr == "@midnight" {
		cronExpr = "0 0 * * *"
	} else if cronExpr == "@hourly" {
		cronExpr = "0 * * * *"
	}

	parts := strings.Fields(cronExpr)
	if len(parts) != 5 {
		return time.Time{}, fmt.Errorf("invalid cron expression (expected 5 fields): %s", cronExpr)
	}

	// Simple forward scanning by minute up to 31 days
	candidate := from.Truncate(time.Minute).Add(time.Minute)
	limit := from.Add(31 * 24 * time.Hour)

	for candidate.Before(limit) {
		if matchField(parts[0], candidate.Minute(), 0, 59) &&
			matchField(parts[1], candidate.Hour(), 0, 23) &&
			matchField(parts[2], candidate.Day(), 1, 31) &&
			matchField(parts[3], int(candidate.Month()), 1, 12) &&
			matchField(parts[4], int(candidate.Weekday()), 0, 6) {
			return candidate, nil
		}
		candidate = candidate.Add(time.Minute)
	}

	return from.Add(24 * time.Hour), nil
}

func matchField(pattern string, val int, min, max int) bool {
	if pattern == "*" {
		return true
	}

	// Handle step values (e.g. */5)
	if strings.HasPrefix(pattern, "*/") {
		stepStr := strings.TrimPrefix(pattern, "*/")
		step, err := strconv.Atoi(stepStr)
		if err == nil && step > 0 {
			return val%step == 0
		}
	}

	// Handle exact numbers
	num, err := strconv.Atoi(pattern)
	if err == nil {
		return num == val
	}

	// Handle comma-separated lists (e.g. 1,15,30)
	if strings.Contains(pattern, ",") {
		items := strings.Split(pattern, ",")
		for _, item := range items {
			if n, err := strconv.Atoi(strings.TrimSpace(item)); err == nil && n == val {
				return true
			}
		}
	}

	return false
}
