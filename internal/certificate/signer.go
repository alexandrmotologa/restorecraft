package certificate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// Signer signs Disaster Recovery Certificates using HMAC-SHA256.
type Signer struct {
	secret         []byte
	signerIdentity string
}

// NewSigner creates a certificate signer with the provided secret key.
func NewSigner(secret string, identity string) *Signer {
	if identity == "" {
		identity = "restorecraft-authority"
	}
	return &Signer{
		secret:         []byte(secret),
		signerIdentity: identity,
	}
}

// CanonicalPayload generates a deterministic string of certificate fields to be signed.
func CanonicalPayload(cert *models.DisasterRecoveryCertificate) string {
	return fmt.Sprintf(
		"cert_id=%s|drill_id=%s|engine=%s|snap_hash=%s|snap_bytes=%d|exec_at=%d|score=%.1f|grade=%s|rto_ms=%d|sla_met=%t|integrity=%t|signer=%s",
		cert.CertificateID,
		cert.DrillID,
		cert.DatabaseEngine,
		cert.SnapshotHash,
		cert.SnapshotBytes,
		cert.ExecutedAt.Unix(),
		cert.HealthScore,
		cert.HealthGrade,
		cert.RtoMs,
		cert.SlaMet,
		cert.IntegrityOk,
		cert.SignerIdentity,
	)
}

// Issue generates and signs a new Disaster Recovery Certificate.
func (s *Signer) Issue(drill *models.DrillRun) (*models.DisasterRecoveryCertificate, error) {
	certID := fmt.Sprintf("cert-%s-%d", drill.ID, time.Now().Unix())

	summary := fmt.Sprintf("Drill '%s' achieved Health Grade %s (Score: %.1f). RTO: %dms (SLA: %ds). Integrity: %t.",
		drill.Name, drill.HealthGrade, drill.HealthScore, drill.Timings.TotalRtoMs, drill.RtoSlaSeconds, drill.Audit.IntegrityOk)

	rtoSlaMs := int64(drill.RtoSlaSeconds) * 1000

	cert := &models.DisasterRecoveryCertificate{
		CertificateID:  certID,
		DrillID:        drill.ID,
		DrillName:      drill.Name,
		DatabaseEngine: drill.Engine,
		SnapshotPath:   drill.SnapshotPath,
		SnapshotHash:   drill.SnapshotHash,
		SnapshotBytes:  drill.SnapshotBytes,
		ExecutedAt:     time.Now().UTC(),
		HealthScore:    drill.HealthScore,
		HealthGrade:    drill.HealthGrade,
		RtoMs:          drill.Timings.TotalRtoMs,
		RtoSlaMs:       rtoSlaMs,
		SlaMet:         drill.SlaMet,
		IntegrityOk:    drill.Audit.IntegrityOk,
		AuditSummary:   summary,
		SignerIdentity: s.signerIdentity,
	}

	payload := CanonicalPayload(cert)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	cert.Signature = hex.EncodeToString(mac.Sum(nil))

	return cert, nil
}
