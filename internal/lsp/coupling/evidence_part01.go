package coupling

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func validateEnvelope(envelope Envelope) error {
	if envelope.Schema != SchemaVersion {
		return fmt.Errorf("schema is %q, want %q", envelope.Schema, SchemaVersion)
	}
	for name, digest := range map[string]string{
		"snapshot": envelope.SnapshotDigest, "registry": envelope.RegistryDigest,
		"toolchain": envelope.ToolchainDigest, "profile": envelope.ProfileDigest,
		"detector result": envelope.DetectorResultDigest, "oracle result": envelope.OracleResultDigest,
		"evidence": envelope.EvidenceDigest,
	} {
		if err := validateDigest(digest); err != nil {
			return fmt.Errorf("%s digest: %w", name, err)
		}
	}
	if !exactText(envelope.Document.URI) || envelope.Document.Version <= 0 {
		return fmt.Errorf("document URI and positive version are required")
	}
	if !envelope.Status.valid() {
		return fmt.Errorf("invalid envelope status %q", envelope.Status)
	}
	if envelope.Status == OutcomePass && envelope.Reason != "" {
		return fmt.Errorf("PASS envelope cannot have a reason")
	}
	if envelope.Status != OutcomePass && !envelope.Reason.valid() {
		return fmt.Errorf("non-PASS envelope requires a known reason")
	}
	for index, explanation := range envelope.Explanations {
		if err := validateExplanation(envelope, explanation); err != nil {
			return fmt.Errorf("explanation %d: %w", index, err)
		}
	}
	expected, err := ComputeEvidenceDigest(envelope)
	if err != nil {
		return fmt.Errorf("evidence digest: %w", err)
	}
	if envelope.EvidenceDigest != expected {
		return fmt.Errorf("evidence digest mismatch")
	}
	return nil
}
func validateDigest(value string) error {
	if len(value) != hex.EncodedLen(32) || strings.TrimSpace(value) != value || strings.ToLower(value) != value {
		return fmt.Errorf("must be a canonical SHA-256 hex digest")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("must be a canonical SHA-256 hex digest")
	}
	return nil
}
