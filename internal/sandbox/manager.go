package sandbox

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// ActiveSandboxInfo describes a running or held sandbox available for debugging.
type ActiveSandboxInfo struct {
	ID        string    `json:"id"`
	DrillID   string    `json:"drill_id"`
	Engine    string    `json:"engine"`
	Type      string    `json:"type"`
	HoldUntil time.Time `json:"hold_until"`
	FilePath  string    `json:"file_path,omitempty"`
}

// QueryResult holds tabular data returned from a read-only debug query.
type QueryResult struct {
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
	RowCount int        `json:"row_count"`
	DurationMs int64    `json:"duration_ms"`
}

// Manager tracks live sandboxes that are held open for manual inspection.
type Manager struct {
	sandboxes map[string]*heldSandbox
	mu        sync.RWMutex
}

type heldSandbox struct {
	info      ActiveSandboxInfo
	sb        Sandbox
	dbPath    string
	dbHandle  *sql.DB
	expiresAt time.Time
}

var globalManager = &Manager{
	sandboxes: make(map[string]*heldSandbox),
}

// GetManager returns the singleton sandbox manager.
func GetManager() *Manager {
	return globalManager
}

// RegisterHeldSandbox keeps a sandbox alive until expiresAt.
func (m *Manager) RegisterHeldSandbox(drillID, engine string, sb Sandbox, duration time.Duration) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	sandboxID := fmt.Sprintf("sb-%s", drillID)
	expiresAt := time.Now().Add(duration)
	details := sb.ConnectionDetails()

	var dbHandle *sql.DB
	if engine == "sqlite" && details.FilePath != "" {
		dbHandle, _ = sql.Open("sqlite", details.FilePath+"?_pragma=busy_timeout(5000)")
	}

	held := &heldSandbox{
		info: ActiveSandboxInfo{
			ID:        sandboxID,
			DrillID:   drillID,
			Engine:    engine,
			Type:      sb.Type(),
			HoldUntil: expiresAt,
			FilePath:  details.FilePath,
		},
		sb:        sb,
		dbPath:    details.FilePath,
		dbHandle:  dbHandle,
		expiresAt: expiresAt,
	}

	m.sandboxes[sandboxID] = held

	// Schedule automatic pruning after expiry
	go func() {
		time.Sleep(duration)
		_ = m.DestroySandbox(sandboxID)
	}()

	return sandboxID
}

// ListSandboxes returns active held sandboxes.
func (m *Manager) ListSandboxes() []ActiveSandboxInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]ActiveSandboxInfo, 0, len(m.sandboxes))
	for _, s := range m.sandboxes {
		res = append(res, s.info)
	}
	return res
}

// QuerySandbox executes a read-only SQL query against an active sandbox.
func (m *Manager) QuerySandbox(ctx context.Context, sandboxID string, query string) (QueryResult, error) {
	m.mu.RLock()
	held, exists := m.sandboxes[sandboxID]
	m.mu.RUnlock()

	if !exists {
		return QueryResult{}, fmt.Errorf("active sandbox '%s' not found or expired", sandboxID)
	}

	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "PRAGMA") && !strings.HasPrefix(trimmed, "EXPLAIN") {
		return QueryResult{}, fmt.Errorf("only read-only statements (SELECT, PRAGMA, EXPLAIN) are permitted in debug console")
	}

	if held.dbHandle == nil {
		return QueryResult{}, fmt.Errorf("direct query not supported for this sandbox driver")
	}

	start := time.Now()
	rows, err := held.dbHandle.QueryContext(ctx, query)
	if err != nil {
		return QueryResult{}, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return QueryResult{}, err
	}

	var resultRows [][]string
	for rows.Next() {
		// Read generic values as strings
		vals := make([]any, len(cols))
		valPtrs := make([]any, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}

		if err := rows.Scan(valPtrs...); err != nil {
			continue
		}

		rowStrs := make([]string, len(cols))
		for i, v := range vals {
			if v == nil {
				rowStrs[i] = "NULL"
			} else if b, ok := v.([]byte); ok {
				rowStrs[i] = string(b)
			} else {
				rowStrs[i] = fmt.Sprintf("%v", v)
			}
		}
		resultRows = append(resultRows, rowStrs)

		// Limit to 200 rows max in debug console
		if len(resultRows) >= 200 {
			break
		}
	}

	return QueryResult{
		Columns:    cols,
		Rows:       resultRows,
		RowCount:   len(resultRows),
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// DestroySandbox terminates and prunes a held sandbox immediately.
func (m *Manager) DestroySandbox(sandboxID string) error {
	m.mu.Lock()
	held, exists := m.sandboxes[sandboxID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	delete(m.sandboxes, sandboxID)
	m.mu.Unlock()

	if held.dbHandle != nil {
		_ = held.dbHandle.Close()
	}

	return held.sb.Cleanup(context.Background())
}
