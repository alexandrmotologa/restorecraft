package certificate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/alexandrmotologa/restorecraft/internal/models"
)

// VerificationResult holds the cryptographic audit result for a certificate.
type VerificationResult struct {
	Valid          bool   `json:"valid"`
	CertificateID  string `json:"certificate_id"`
	SignerIdentity string `json:"signer_identity"`
	Reason         string `json:"reason,omitempty"`
}

// VerifyCertificate checks if the signature of a certificate matches using the secret key.
func VerifyCertificate(cert *models.DisasterRecoveryCertificate, secret string) VerificationResult {
	if cert == nil {
		return VerificationResult{Valid: false, Reason: "nil certificate"}
	}
	if cert.Signature == "" {
		return VerificationResult{Valid: false, CertificateID: cert.CertificateID, Reason: "missing signature"}
	}

	expectedSig, err := hex.DecodeString(cert.Signature)
	if err != nil {
		return VerificationResult{Valid: false, CertificateID: cert.CertificateID, Reason: "malformed hex signature"}
	}

	payload := CanonicalPayload(cert)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	calculatedSig := mac.Sum(nil)

	if !hmac.Equal(expectedSig, calculatedSig) {
		return VerificationResult{
			Valid:          false,
			CertificateID:  cert.CertificateID,
			SignerIdentity: cert.SignerIdentity,
			Reason:         "digital signature mismatch (certificate payload altered or invalid secret key)",
		}
	}

	return VerificationResult{
		Valid:          true,
		CertificateID:  cert.CertificateID,
		SignerIdentity: cert.SignerIdentity,
	}
}

// VerifyCertificateBytes verifies a certificate serialized as JSON bytes.
func VerifyCertificateBytes(data []byte, secret string) (VerificationResult, error) {
	var cert models.DisasterRecoveryCertificate
	if err := json.Unmarshal(data, &cert); err != nil {
		return VerificationResult{Valid: false, Reason: fmt.Sprintf("unmarshal error: %v", err)}, err
	}
	return VerifyCertificate(&cert, secret), nil
}
