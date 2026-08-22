package artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func ValidateFoundationBaseline(raw []byte, receipt Receipt) error {
	reference := FoundationBaseline()
	if digestBytes(raw) != reference.FileSHA256 {
		return fmt.Errorf("FAIL_CLOSED: readiness baseline file digest mismatch")
	}
	if err := Validate(receipt); err != nil {
		return fmt.Errorf("FAIL_CLOSED: readiness baseline invalid: %w", err)
	}
	summary := receipt.Snapshot.Summary
	switch {
	case receipt.HeadSHA != reference.HeadSHA:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline head mismatch")
	case receipt.ArtifactDigest != reference.ArtifactDigest:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline artifact mismatch")
	case receipt.Snapshot.Digest != reference.SnapshotDigest:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline snapshot mismatch")
	case receipt.Snapshot.RegistryDigest != reference.RegistryDigest:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline registry mismatch")
	case summary.Completed != reference.Completed || summary.Total != reference.Total:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline ratio mismatch")
	case summary.ReadinessBPS != reference.BasisPoints:
		return fmt.Errorf("FAIL_CLOSED: readiness baseline basis points mismatch")
	}
	return nil
}

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
