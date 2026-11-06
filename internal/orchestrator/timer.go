package orchestrator

import (
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// PhaseTimer measures nanosecond durations across drill phases.
type PhaseTimer struct {
	startedAt time.Time
	timings   models.PhaseTimings
}

// NewPhaseTimer initializes a stopwatch.
func NewPhaseTimer() *PhaseTimer {
	return &PhaseTimer{
		startedAt: time.Now(),
	}
}

// Measure executes fn and records its execution time in milliseconds into target.
func (p *PhaseTimer) Measure(target *int64, fn func() error) error {
	start := time.Now()
	err := fn()
	*target = time.Since(start).Milliseconds()
	return err
}

// Complete finalizes total RTO.
func (p *PhaseTimer) Complete() models.PhaseTimings {
	p.timings.TotalRtoMs = time.Since(p.startedAt).Milliseconds()
	return p.timings
}

// CurrentTimings returns current timing values.
func (p *PhaseTimer) CurrentTimings() models.PhaseTimings {
	t := p.timings
	t.TotalRtoMs = time.Since(p.startedAt).Milliseconds()
	return t
}
