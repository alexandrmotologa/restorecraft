package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/orchestrator"
	"github.com/alexandrmotologa/restorecraft/internal/store"
)

func setupTestServer(t *testing.T) (*Server, *store.Store, string) {
	tempDir, err := os.MkdirTemp("", "restorecraft_server_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}

	fixturesDir := filepath.Join(tempDir, "fixtures")
	if err := fixtures.GenerateAll(fixturesDir); err != nil {
		t.Fatalf("fixtures: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Security.SigningSecret = "server-test-secret-32-chars-long"
	cfg.Storage.Local.BaseDir = fixturesDir

	st, err := store.NewStore(filepath.Join(tempDir, "data"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	orch := orchestrator.NewOrchestrator(cfg, st)
	srv := NewServer(cfg, st, orch, nil)

	return srv, st, tempDir
}

func TestServerStatusAndConfig(t *testing.T) {
	srv, st, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)
	defer st.Close()

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var status models.SystemStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("unmarshal status: %v", err)
	}

	reqCfg := httptest.NewRequest("GET", "/api/config", nil)
	wCfg := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wCfg, reqCfg)

	if wCfg.Code != http.StatusOK {
		t.Fatalf("expected config status 200, got %d", wCfg.Code)
	}
}

func TestServerRunDrillAndVerifyCertificate(t *testing.T) {
	srv, st, tempDir := setupTestServer(t)
	defer os.RemoveAll(tempDir)
	defer st.Close()

	runReq := RunDrillRequest{
		Engine:        "sqlite",
		SnapshotPath:  filepath.Join(tempDir, "fixtures", "valid_ecommerce.sqlite"),
		StorageType:   "local",
		RtoSlaSeconds: 60,
	}

	bodyBytes, _ := json.Marshal(runReq)
	req := httptest.NewRequest("POST", "/api/drills/run?sync=true", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("run drill failed with status %d: %s", w.Code, w.Body.String())
	}

	var run models.DrillRun
	if err := json.Unmarshal(w.Body.Bytes(), &run); err != nil {
		t.Fatalf("unmarshal run: %v", err)
	}

	if run.Status != models.StatusCompleted {
		t.Fatalf("expected status completed, got %s", run.Status)
	}
	if run.HealthGrade != models.GradeAPlus {
		t.Fatalf("expected Grade A+, got %s", run.HealthGrade)
	}
	if run.Certificate == nil {
		t.Fatalf("expected certificate to be issued")
	}

	// Now test certificate verification route
	verifyPayload := map[string]any{
		"certificate": run.Certificate,
		"secret_key":  "server-test-secret-32-chars-long",
	}
	verifyBytes, _ := json.Marshal(verifyPayload)

	vReq := httptest.NewRequest("POST", "/api/certificates/verify", bytes.NewBuffer(verifyBytes))
	vReq.Header.Set("Content-Type", "application/json")
	vW := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(vW, vReq)

	if vW.Code != http.StatusOK {
		t.Fatalf("verify failed with status %d: %s", vW.Code, vW.Body.String())
	}

	var verifyRes map[string]any
	if err := json.Unmarshal(vW.Body.Bytes(), &verifyRes); err != nil {
		t.Fatalf("unmarshal verify response: %v", err)
	}

	if verifyRes["valid"] != true {
		t.Fatalf("expected valid certificate, got: %+v", verifyRes)
	}

	// Test Prometheus metrics endpoint
	mReq := httptest.NewRequest("GET", "/metrics", nil)
	mW := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(mW, mReq)
	if mW.Code != http.StatusOK {
		t.Fatalf("metrics endpoint returned status %d", mW.Code)
	}
	if !bytes.Contains(mW.Body.Bytes(), []byte("restorecraft_up 1")) {
		t.Fatalf("expected restorecraft_up metric")
	}

	// Test sandboxes endpoint
	sReq := httptest.NewRequest("GET", "/api/sandboxes", nil)
	sW := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(sW, sReq)
	if sW.Code != http.StatusOK {
		t.Fatalf("sandboxes endpoint returned status %d", sW.Code)
	}

	// Test bisect endpoint
	bisectPayload := map[string]string{
		"directory": filepath.Join(tempDir, "fixtures"),
		"engine":    "sqlite",
	}
	bBytes, _ := json.Marshal(bisectPayload)
	bReq := httptest.NewRequest("POST", "/api/drills/bisect", bytes.NewBuffer(bBytes))
	bReq.Header.Set("Content-Type", "application/json")
	bW := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(bW, bReq)
	if bW.Code != http.StatusOK {
		t.Fatalf("bisect endpoint returned status %d: %s", bW.Code, bW.Body.String())
	}
}
