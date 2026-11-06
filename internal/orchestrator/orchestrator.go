package orchestrator

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/auditor"
	"github.com/alexandrmotologa/restorecraft/internal/certificate"
	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/engine"
	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/notifier"
	"github.com/alexandrmotologa/restorecraft/internal/sandbox"
	"github.com/alexandrmotologa/restorecraft/internal/storage"
	"github.com/alexandrmotologa/restorecraft/internal/store"
)

// LogHub broadcasts real-time log events to SSE subscribers.
type LogHub struct {
	subscribers map[chan models.LogEntry]bool
	mu          sync.RWMutex
}

func NewLogHub() *LogHub {
	return &LogHub{
		subscribers: make(map[chan models.LogEntry]bool),
	}
}

func (h *LogHub) Subscribe() (chan models.LogEntry, func()) {
	ch := make(chan models.LogEntry, 100)
	h.mu.Lock()
	h.subscribers[ch] = true
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		close(ch)
		h.mu.Unlock()
	}
	return ch, unsubscribe
}

func (h *LogHub) Broadcast(entry models.LogEntry) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subscribers {
		select {
		case ch <- entry:
		default:
		}
	}
}

// Orchestrator coordinates disaster recovery drill execution.
type Orchestrator struct {
	cfg      *config.Config
	store    *store.Store
	auditor  *auditor.Auditor
	signer   *certificate.Signer
	notifier *notifier.Dispatcher
	logHub   *LogHub
	mu       sync.Mutex
}

// NewOrchestrator creates a new drill orchestrator.
func NewOrchestrator(cfg *config.Config, s *store.Store) *Orchestrator {
	return &Orchestrator{
		cfg:      cfg,
		store:    s,
		auditor:  auditor.NewAuditor(),
		signer:   certificate.NewSigner(cfg.Security.SigningSecret, cfg.Security.SignerIdentity),
		notifier: notifier.NewDispatcher(cfg.Notifications),
		logHub:   NewLogHub(),
	}
}

func (o *Orchestrator) SubscribeLogs() (chan models.LogEntry, func()) {
	return o.logHub.Subscribe()
}

func (o *Orchestrator) emitLog(drillID, level, phase, msg string) {
	entry := models.LogEntry{
		DrillID:   drillID,
		Timestamp: time.Now().UTC(),
		Level:     level,
		Phase:     phase,
		Message:   msg,
	}
	o.logHub.Broadcast(entry)
}

// RunDrill executes a complete automated disaster recovery drill.
func (o *Orchestrator) RunDrill(ctx context.Context, drillCfg config.DrillConfig) (*models.DrillRun, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	drillID := fmt.Sprintf("drill-%s-%d", drillCfg.ID, time.Now().Unix())
	startTime := time.Now().UTC()
	timer := NewPhaseTimer()

	slaSeconds := drillCfg.RtoSlaSeconds
	if slaSeconds <= 0 {
		slaSeconds = o.cfg.Defaults.RtoSlaSeconds
	}
	if slaSeconds <= 0 {
		slaSeconds = 300
	}

	drill := &models.DrillRun{
		ID:             drillID,
		ConfigID:       drillCfg.ID,
		Name:           drillCfg.Name,
		Engine:         drillCfg.Engine,
		Status:         models.StatusRunning,
		SnapshotSource: drillCfg.StorageType,
		SnapshotPath:   drillCfg.SnapshotPath,
		StartedAt:      startTime,
		RtoSlaSeconds:  slaSeconds,
	}

	// Persist initial running state
	_ = o.store.SaveDrill(drill)

	o.emitLog(drillID, "STEP", "init", fmt.Sprintf("Starting disaster recovery drill '%s' (Engine: %s, SLA: %ds)", drillCfg.Name, drillCfg.Engine, slaSeconds))

	// 1. Storage Provider resolution
	var storageProvider storage.Provider
	if drillCfg.StorageType == "s3" && o.cfg.Storage.S3.Enabled {
		s3Prov, err := storage.NewS3Provider(ctx, o.cfg.Storage.S3)
		if err != nil {
			return o.failDrill(drill, timer, "init", fmt.Sprintf("initialize s3 provider: %v", err))
		}
		storageProvider = s3Prov
	} else {
		baseDir := o.cfg.Storage.Local.BaseDir
		storageProvider = storage.NewLocalProvider(baseDir)
	}

	// 2. Fetch snapshot and measure fetch time
	o.emitLog(drillID, "INFO", "fetch", fmt.Sprintf("Acquiring snapshot from %s: %s", storageProvider.Type(), drillCfg.SnapshotPath))
	var rawStream io.ReadCloser
	var snapshotBytes int64

	fetchErr := timer.Measure(&drill.Timings.FetchMs, func() error {
		var err error
		rawStream, snapshotBytes, err = storageProvider.Fetch(ctx, drillCfg.SnapshotPath)
		return err
	})

	if fetchErr != nil {
		return o.failDrill(drill, timer, "fetch", fmt.Sprintf("fetch snapshot failed: %v", fetchErr))
	}
	defer rawStream.Close()

	drill.SnapshotBytes = snapshotBytes
	o.emitLog(drillID, "INFO", "fetch", fmt.Sprintf("Acquired %d bytes in %dms", snapshotBytes, drill.Timings.FetchMs))

	// Wrap stream in HashingReader to calculate SHA-256 on the fly
	hashingReader := storage.NewHashingReader(rawStream)

	// 3. Decompress snapshot stream
	o.emitLog(drillID, "INFO", "decompress", "Preparing decompression stream")
	var decompressedStream io.Reader
	var compType string

	decompressErr := timer.Measure(&drill.Timings.DecompressMs, func() error {
		var err error
		decompressedStream, compType, err = storage.OpenDecompressedStream(hashingReader, drillCfg.SnapshotPath)
		return err
	})

	if decompressErr != nil {
		return o.failDrill(drill, timer, "decompress", fmt.Sprintf("decompression error: %v", decompressErr))
	}
	o.emitLog(drillID, "INFO", "decompress", fmt.Sprintf("Detected compression: %s (Duration: %dms)", compType, drill.Timings.DecompressMs))

	// 4. Provision ephemeral sandbox
	o.emitLog(drillID, "STEP", "sandbox", fmt.Sprintf("Provisioning ephemeral isolated sandbox for %s", drillCfg.Engine))
	var sb sandbox.Sandbox

	if drillCfg.Engine == "postgres" || drillCfg.Engine == "mysql" {
		sb = sandbox.NewDockerSandbox(drillID, drillCfg.Engine, drillCfg.Docker)
	} else {
		// Native SQLite sandbox
		sb = sandbox.NewNativeSandbox(fmt.Sprintf("%s.db", drillID))
	}

	sandboxErr := timer.Measure(&drill.Timings.SandboxMs, func() error {
		return sb.Provision(ctx)
	})

	if sandboxErr != nil {
		return o.failDrill(drill, timer, "sandbox", fmt.Sprintf("sandbox provisioning failed: %v", sandboxErr))
	}
	defer func() {
		o.emitLog(drillID, "INFO", "teardown", "Pruning ephemeral sandbox and temporary volumes")
		_ = timer.Measure(&drill.Timings.TeardownMs, func() error {
			return sb.Cleanup(context.Background())
		})
		o.emitLog(drillID, "INFO", "teardown", fmt.Sprintf("Teardown completed in %dms", drill.Timings.TeardownMs))
	}()

	o.emitLog(drillID, "INFO", "sandbox", fmt.Sprintf("Sandbox ready in %dms (Driver: %s)", drill.Timings.SandboxMs, sb.Type()))

	// 5. Initialize database engine driver
	var dbEngine engine.Engine
	connDetails := sb.ConnectionDetails()

	switch drillCfg.Engine {
	case "sqlite":
		dbEngine = engine.NewSQLiteEngine(connDetails.FilePath)
	case "postgres":
		dbEngine = engine.NewPostgresEngine(connDetails)
	case "mysql":
		dbEngine = engine.NewMySQLEngine(connDetails)
	default:
		dbEngine = engine.NewSQLiteEngine(connDetails.FilePath)
	}
	defer dbEngine.Close()

	// 6. Restore snapshot stream into sandbox
	o.emitLog(drillID, "STEP", "restore", "Restoring snapshot into database...")
	restoreErr := timer.Measure(&drill.Timings.RestoreMs, func() error {
		return dbEngine.RestoreSnapshot(ctx, decompressedStream, filepath.Base(drillCfg.SnapshotPath))
	})

	// Record hash computed on the fly
	drill.SnapshotHash = hashingReader.SumHex()

	if restoreErr != nil {
		return o.failDrill(drill, timer, "restore", fmt.Sprintf("restore error: %v", restoreErr))
	}
	o.emitLog(drillID, "INFO", "restore", fmt.Sprintf("Database restored successfully in %dms (SHA-256: %s)", drill.Timings.RestoreMs, drill.SnapshotHash[:16]+"..."))

	// 7. Audit structural integrity and constraints
	o.emitLog(drillID, "STEP", "audit", "Executing deep integrity and parity audit...")
	var auditReport models.AuditReport
	var healthScore float64
	var healthGrade models.HealthGrade

	auditErr := timer.Measure(&drill.Timings.AuditMs, func() error {
		var err error
		auditReport, healthScore, healthGrade, err = o.auditor.Audit(
			ctx,
			dbEngine,
			drillCfg.Parity,
			timer.CurrentTimings().TotalRtoMs,
			slaSeconds,
		)
		return err
	})

	if auditErr != nil {
		return o.failDrill(drill, timer, "audit", fmt.Sprintf("audit execution error: %v", auditErr))
	}

	drill.Audit = auditReport
	drill.HealthScore = healthScore
	drill.HealthGrade = healthGrade

	// Complete stopwatch
	timings := timer.Complete()
	drill.Timings = timings

	// Verify SLA
	slaMet := timings.TotalRtoMs <= int64(slaSeconds)*1000
	drill.SlaMet = slaMet
	drill.Status = models.StatusCompleted
	now := time.Now().UTC()
	drill.CompletedAt = &now

	o.emitLog(drillID, "STEP", "audit", fmt.Sprintf("Audit complete. Health Grade: %s (Score: %.1f). Total RTO: %dms (SLA: %ds, Met: %t)",
		healthGrade, healthScore, timings.TotalRtoMs, slaSeconds, slaMet))

	// 8. Issue cryptographic Disaster Recovery Certificate
	cert, err := o.signer.Issue(drill)
	if err == nil {
		drill.Certificate = cert
		o.emitLog(drillID, "INFO", "cert", fmt.Sprintf("Issued Disaster Recovery Certificate: %s", cert.CertificateID))
	}

	// 9. Persist final drill state
	_ = o.store.SaveDrill(drill)

	// 10. Dispatch alerts
	o.notifier.Dispatch(ctx, drill)

	o.emitLog(drillID, "STEP", "done", fmt.Sprintf("Disaster recovery drill %s completed with grade %s", drill.ID, drill.HealthGrade))

	return drill, nil
}

func (o *Orchestrator) failDrill(drill *models.DrillRun, timer *PhaseTimer, phase, errMsg string) (*models.DrillRun, error) {
	drill.Status = models.StatusFailed
	drill.ErrorMessage = errMsg
	drill.HealthScore = 0.0
	drill.HealthGrade = models.GradeF
	drill.SlaMet = false
	drill.Timings = timer.Complete()
	now := time.Now().UTC()
	drill.CompletedAt = &now

	o.emitLog(drill.ID, "ERROR", phase, fmt.Sprintf("FAILED: %s", errMsg))

	_ = o.store.SaveDrill(drill)
	o.notifier.Dispatch(context.Background(), drill)

	return drill, fmt.Errorf("%s: %s", phase, errMsg)
}

// RunDrillByID finds the drill in configuration and executes it.
func (o *Orchestrator) RunDrillByID(ctx context.Context, id string) (*models.DrillRun, error) {
	for _, d := range o.cfg.Drills {
		if d.ID == id {
			return o.RunDrill(ctx, d)
		}
	}
	return nil, fmt.Errorf("drill config '%s' not found", id)
}
