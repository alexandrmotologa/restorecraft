package certificate

import (
	"testing"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

func TestSigningAndVerification(t *testing.T) {
	secret := "test-secret-key-high-entropy-12345"
	signer := NewSigner(secret, "test-signer-authority")

	now := time.Now().UTC()
	drill := &models.DrillRun{
		ID:             "drill-sign-test",
		Name:           "Test Sign Drill",
		Engine:         "sqlite",
		SnapshotPath:   "test.sqlite",
		SnapshotHash:   "abcdef1234567890",
		SnapshotBytes:  2048,
		StartedAt:      now,
		HealthScore:    95.0,
		HealthGrade:    models.GradeAPlus,
		RtoSlaSeconds:  60,
		SlaMet:         true,
		Timings: models.PhaseTimings{
			TotalRtoMs: 850,
		},
		Audit: models.AuditReport{
			IntegrityOk: true,
		},
	}

	cert, err := signer.Issue(drill)
	if err != nil {
		t.Fatalf("signer.Issue: %v", err)
	}

	if cert.Signature == "" {
		t.Fatalf("expected non-empty signature")
	}

	// Verify with correct secret
	res := VerifyCertificate(cert, secret)
	if !res.Valid {
		t.Fatalf("expected valid certificate verification, got: %s", res.Reason)
	}

	// Verify with incorrect secret
	resWrong := VerifyCertificate(cert, "wrong-secret-key")
	if resWrong.Valid {
		t.Fatalf("expected verification failure with incorrect secret")
	}

	// Verify tampering detection
	cert.HealthScore = 50.0 // Tamper with score
	resTampered := VerifyCertificate(cert, secret)
	if resTampered.Valid {
		t.Fatalf("expected verification failure on tampered certificate")
	}
}
