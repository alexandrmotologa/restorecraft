package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/certificate"
	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/alexandrmotologa/restorecraft/internal/fixtures"
	"github.com/alexandrmotologa/restorecraft/internal/models"
	"github.com/alexandrmotologa/restorecraft/internal/orchestrator"
	"github.com/alexandrmotologa/restorecraft/internal/server"
	"github.com/alexandrmotologa/restorecraft/internal/store"
	"github.com/alexandrmotologa/restorecraft/web"
)

var (
	Version   = "1.0.0"
	BuildDate = "2026-11-06"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "drill":
		cmdDrill(os.Args[2:])
	case "verify-cert":
		cmdVerifyCert(os.Args[2:])
	case "fixtures":
		cmdFixtures(os.Args[2:])
	case "version":
		fmt.Printf("restorecraft version %s (built %s)\n", Version, BuildDate)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`restorecraft: Automated Database Disaster Recovery Drill & Parity Engine

Usage:
  restorecraft <command> [arguments]

Available Commands:
  serve         Start background daemon and embedded React 19 Recovery Cockpit
  drill         Execute an on-demand disaster recovery drill via CLI
  verify-cert   Verify the cryptographic HMAC-SHA256 signature of a recovery certificate
  fixtures      Generate valid and deliberately corrupted backup testbeds
  version       Print software version and build info

Run 'restorecraft <command> -h' for more details on a specific command.`)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 0, "HTTP port for web cockpit (overrides config)")
	host := fs.String("host", "", "HTTP host interface (overrides config)")
	cfgPath := fs.String("config", "", "Path to restorecraft.yaml config file")
	dataDir := fs.String("data-dir", "", "Path to persistence directory")
	_ = fs.Parse(args)

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *dataDir != "" {
		cfg.Server.DataDir = *dataDir
	}

	// Auto-generate test fixtures if none exist
	if _, err := os.Stat("./test-backups/valid_ecommerce.sqlite"); os.IsNotExist(err) {
		_ = fixtures.GenerateAll("./test-backups")
	}

	st, err := store.NewStore(cfg.Server.DataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize store: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	orch := orchestrator.NewOrchestrator(cfg, st)
	srv := server.NewServer(cfg, st, orch, web.GetFS())

	fmt.Printf("====================================================\n")
	fmt.Printf("  restorecraft Recovery Cockpit v%s\n", Version)
	fmt.Printf("  Serving on http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  Persistence: %s\n", cfg.Server.DataDir)
	fmt.Printf("====================================================\n")

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-stop
	fmt.Println("\nShutting down restorecraft daemon...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	fmt.Println("Shutdown completed cleanly.")
}

func cmdDrill(args []string) {
	fs := flag.NewFlagSet("drill", flag.ExitOnError)
	engine := fs.String("engine", "sqlite", "Target database engine: sqlite, postgres, mysql")
	snapshot := fs.String("snapshot", "./test-backups/valid_ecommerce.sqlite", "Snapshot path or S3 key")
	storageType := fs.String("storage", "local", "Storage type: local, s3")
	sla := fs.Int("sla", 60, "RTO SLA threshold in seconds")
	cfgPath := fs.String("config", "", "Config path")
	dataDir := fs.String("data-dir", "./data", "Data directory")
	_ = fs.Parse(args)

	cfg, _ := config.LoadConfig(*cfgPath)
	if *dataDir != "" {
		cfg.Server.DataDir = *dataDir
	}

	// Ensure test fixtures exist if target is default
	if *snapshot == "./test-backups/valid_ecommerce.sqlite" {
		if _, err := os.Stat(*snapshot); os.IsNotExist(err) {
			_ = fixtures.GenerateAll("./test-backups")
		}
	}

	st, err := store.NewStore(cfg.Server.DataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init store: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	orch := orchestrator.NewOrchestrator(cfg, st)

	drillCfg := config.DrillConfig{
		ID:            "cli-drill",
		Name:          fmt.Sprintf("CLI On-Demand Drill (%s)", *engine),
		Engine:        *engine,
		StorageType:   *storageType,
		SnapshotPath:  *snapshot,
		RtoSlaSeconds: *sla,
		Parity: config.ParityConfig{
			ExpectedTables: []string{"users", "orders"},
		},
	}

	fmt.Printf("Executing disaster recovery drill for %s snapshot: %s\n", *engine, *snapshot)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	run, err := orch.RunDrill(ctx, drillCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n[DRILL FAILED]: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n--- Drill Results ---\n")
	fmt.Printf("Status:       %s\n", run.Status)
	fmt.Printf("Health Grade: %s (Score: %.1f)\n", run.HealthGrade, run.HealthScore)
	fmt.Printf("Total RTO:    %dms (SLA: %ds, Met: %t)\n", run.Timings.TotalRtoMs, run.RtoSlaSeconds, run.SlaMet)
	fmt.Printf("Phase Timings:\n")
	fmt.Printf("  - Fetch:       %dms\n", run.Timings.FetchMs)
	fmt.Printf("  - Decompress:  %dms\n", run.Timings.DecompressMs)
	fmt.Printf("  - Sandbox:     %dms\n", run.Timings.SandboxMs)
	fmt.Printf("  - Restore:     %dms\n", run.Timings.RestoreMs)
	fmt.Printf("  - Audit:       %dms\n", run.Timings.AuditMs)
	fmt.Printf("  - Teardown:    %dms\n", run.Timings.TeardownMs)
	fmt.Printf("Snapshot SHA: %s\n", run.SnapshotHash)
	fmt.Printf("Integrity OK: %t\n", run.Audit.IntegrityOk)

	if run.Certificate != nil {
		fmt.Printf("Certificate:  %s (Signed by %s)\n", run.Certificate.CertificateID, run.Certificate.SignerIdentity)
	}

	if run.HealthGrade == "F" || run.Status == "failed" {
		os.Exit(1)
	}
}

func cmdVerifyCert(args []string) {
	fs := flag.NewFlagSet("verify-cert", flag.ExitOnError)
	certPath := fs.String("cert", "", "Path to Disaster Recovery Certificate JSON file")
	secret := fs.String("secret", "", "Authority HMAC-SHA256 secret key")
	_ = fs.Parse(args)

	if *certPath == "" {
		fmt.Fprintln(os.Stderr, "error: --cert flag is required")
		os.Exit(1)
	}

	data, err := os.ReadFile(*certPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read certificate: %v\n", err)
		os.Exit(1)
	}

	if *secret == "" {
		*secret = os.Getenv("RESTORECRAFT_SIGNING_SECRET")
		if *secret == "" {
			*secret = "restorecraft-default-development-secret-key-32b"
		}
	}

	var cert models.DisasterRecoveryCertificate
	if err := json.Unmarshal(data, &cert); err != nil {
		fmt.Fprintf(os.Stderr, "parse certificate json: %v\n", err)
		os.Exit(1)
	}

	res := certificate.VerifyCertificate(&cert, *secret)
	if res.Valid {
		fmt.Printf("VALID: Certificate %s is cryptographically verified!\n", cert.CertificateID)
		fmt.Printf("Issued for:   %s (%s)\n", cert.DrillName, cert.DatabaseEngine)
		fmt.Printf("Health Grade: %s (Score: %.1f)\n", cert.HealthGrade, cert.HealthScore)
		fmt.Printf("RTO:          %dms (SLA: %dms, Met: %t)\n", cert.RtoMs, cert.RtoSlaMs, cert.SlaMet)
		fmt.Printf("Signer:       %s\n", cert.SignerIdentity)
		fmt.Printf("Signature:    %s\n", cert.Signature)
	} else {
		fmt.Printf("INVALID: Certificate verification failed!\n")
		fmt.Printf("Reason: %s\n", res.Reason)
		os.Exit(1)
	}
}

func cmdFixtures(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: restorecraft fixtures generate [--dir ./test-backups]")
		os.Exit(1)
	}

	if args[0] == "generate" {
		fs := flag.NewFlagSet("generate", flag.ExitOnError)
		dir := fs.String("dir", "./test-backups", "Output directory for test fixtures")
		_ = fs.Parse(args[1:])

		fmt.Printf("Generating synthetic testbeds and corrupted snapshots in %s...\n", *dir)
		if err := fixtures.GenerateAll(*dir); err != nil {
			fmt.Fprintf(os.Stderr, "generate fixtures failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Fixtures created successfully:")
		fmt.Println("  1. valid_ecommerce.sqlite  (Clean SQLite binary)")
		fmt.Println("  2. valid_ecommerce.sql.gz  (Clean gzipped SQL dump)")
		fmt.Println("  3. corrupt_header.sqlite   (Corrupted binary header)")
		fmt.Println("  4. fk_violation.sqlite     (Broken relational foreign keys)")
		fmt.Println("  5. truncated.sql.gz        (Truncated gzip stream)")
	}
}
