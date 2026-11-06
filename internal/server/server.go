package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/certificate"
	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/orchestrator"
	"github.com/alexandrmotologa/restorecraft/internal/store"
)

// Server coordinates the HTTP API, SSE streaming, and embedded frontend UI.
type Server struct {
	cfg          *config.Config
	store        *store.Store
	orchestrator *orchestrator.Orchestrator
	staticFS     fs.FS
	httpServer   *http.Server
}

// NewServer initializes the HTTP server.
func NewServer(cfg *config.Config, s *store.Store, orch *orchestrator.Orchestrator, staticFS fs.FS) *Server {
	srv := &Server{
		cfg:          cfg,
		store:        s,
		orchestrator: orch,
		staticFS:     staticFS,
	}

	mux := http.NewServeMux()
	srv.registerRoutes(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv.httpServer = &http.Server{
		Addr:         addr,
		Handler:      srv.corsMiddleware(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	return srv
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// API Endpoints
	mux.HandleFunc("GET /api/status", s.handleGetStatus)
	mux.HandleFunc("GET /api/drills", s.handleListDrills)
	mux.HandleFunc("GET /api/drills/{id}", s.handleGetDrill)
	mux.HandleFunc("POST /api/drills/run", s.handleRunDrill)
	mux.HandleFunc("GET /api/drills/stream", s.handleLiveLogStream)
	mux.HandleFunc("POST /api/certificates/verify", s.handleVerifyCertificate)
	mux.HandleFunc("POST /api/fixtures/generate", s.handleGenerateFixtures)
	mux.HandleFunc("GET /api/config", s.handleGetConfig)

	// Embedded UI / Static assets
	if s.staticFS != nil {
		fileServer := http.FileServer(http.FS(s.staticFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			// Check if file exists in embedded FS
			f, err := s.staticFS.Open(path)
			if err != nil {
				// Fallback to index.html for React SPA client-side routing
				indexFile, err := s.staticFS.Open("index.html")
				if err != nil {
					http.NotFound(w, r)
					return
				}
				defer indexFile.Close()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.ServeContent(w, r, "index.html", time.Now(), indexFile.(ioReadSeeker))
				return
			}
			defer f.Close()

			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"service": "restorecraft",
				"status":  "running",
				"version": s.cfg.Version,
			})
		})
	}
}

type ioReadSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.store.GetSystemStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleListDrills(w http.ResponseWriter, r *http.Request) {
	drills, err := s.store.ListDrills(50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if drills == nil {
		drills = []models.DrillRun{}
	}
	writeJSON(w, http.StatusOK, drills)
}

func (s *Server) handleGetDrill(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	drill, err := s.store.GetDrill(id)
	if err != nil {
		http.Error(w, "drill not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, drill)
}

type RunDrillRequest struct {
	ConfigID      string `json:"config_id,omitempty"`
	Engine        string `json:"engine,omitempty"`
	SnapshotPath  string `json:"snapshot_path,omitempty"`
	StorageType   string `json:"storage_type,omitempty"`
	RtoSlaSeconds int    `json:"rto_sla_seconds,omitempty"`
}

func (s *Server) handleRunDrill(w http.ResponseWriter, r *http.Request) {
	var req RunDrillRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var drillCfg config.DrillConfig
	if req.ConfigID != "" {
		for _, d := range s.cfg.Drills {
			if d.ID == req.ConfigID {
				drillCfg = d
				break
			}
		}
	}

	if drillCfg.ID == "" {
		drillCfg = config.DrillConfig{
			ID:            "manual-drill",
			Name:          "Manual On-Demand Disaster Recovery Drill",
			Engine:        req.Engine,
			StorageType:   req.StorageType,
			SnapshotPath:  req.SnapshotPath,
			RtoSlaSeconds: req.RtoSlaSeconds,
		}
		if drillCfg.Engine == "" {
			drillCfg.Engine = "sqlite"
		}
		if drillCfg.SnapshotPath == "" {
			drillCfg.SnapshotPath = "./test-backups/valid_ecommerce.sqlite"
		}
		if drillCfg.StorageType == "" {
			drillCfg.StorageType = "local"
		}
		if drillCfg.RtoSlaSeconds <= 0 {
			drillCfg.RtoSlaSeconds = 60
		}
	}

	// Run drill asynchronously or synchronously based on query param
	syncRun := r.URL.Query().Get("sync") == "true"
	if syncRun {
		drill, err := s.orchestrator.RunDrill(r.Context(), drillCfg)
		if err != nil {
			writeJSON(w, http.StatusOK, drill)
			return
		}
		writeJSON(w, http.StatusOK, drill)
		return
	}

	go func() {
		_, _ = s.orchestrator.RunDrill(context.Background(), drillCfg)
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "Drill execution initiated",
		"engine":  drillCfg.Engine,
	})
}

func (s *Server) handleLiveLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsubscribe := s.orchestrator.SubscribeLogs()
	defer unsubscribe()

	// Initial connected ping
	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-ch:
			if !ok {
				return
			}
			bytes, err := json.Marshal(entry)
			if err == nil {
				fmt.Fprintf(w, "event: log\ndata: %s\n\n", string(bytes))
				flusher.Flush()
			}
		}
	}
}

func (s *Server) handleVerifyCertificate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Certificate *models.DisasterRecoveryCertificate `json:"certificate"`
		SecretKey   string                               `json:"secret_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	secret := body.SecretKey
	if secret == "" {
		secret = s.cfg.Security.SigningSecret
	}

	result := certificate.VerifyCertificate(body.Certificate, secret)
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGenerateFixtures(w http.ResponseWriter, r *http.Request) {
	dir := "./test-backups"
	if err := fixtures.GenerateAll(dir); err != nil {
		http.Error(w, fmt.Sprintf("generate fixtures: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Test fixtures generated in ./test-backups",
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	safeConfig := map[string]any{
		"version": s.cfg.Version,
		"server":  s.cfg.Server,
		"drills":  s.cfg.Drills,
		"defaults": s.cfg.Defaults,
	}
	writeJSON(w, http.StatusOK, safeConfig)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Start begins listening on the configured address.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully terminates the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
