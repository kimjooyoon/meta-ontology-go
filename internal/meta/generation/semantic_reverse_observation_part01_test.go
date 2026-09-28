package generation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSemanticOperationReverseObservationBindsAuthorityAndArtifacts(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "self-improvement-minimal-loop", "operation-envelope.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	outputDir := t.TempDir()
	roundTrip, err := GenerateAndReverseObserveSemanticOperationEnvelope(source, "C2", outputDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if roundTrip.Generated.IR.AuthorityDigest != roundTrip.Reverse.AuthorityDigest {
		t.Fatalf("authority digest mismatch: generated %q, reverse %q", roundTrip.Generated.IR.AuthorityDigest, roundTrip.Reverse.AuthorityDigest)
	}
	if roundTrip.Reverse.Decision != roundTrip.Generated.Receipt.Decision.Decision || roundTrip.Reverse.Reason != roundTrip.Generated.Receipt.Decision.Reason {
		t.Fatalf("decision mismatch: generated %s/%s, reverse %s/%s", roundTrip.Generated.Receipt.Decision.Decision, roundTrip.Generated.Receipt.Decision.Reason, roundTrip.Reverse.Decision, roundTrip.Reverse.Reason)
	}
	if roundTrip.Reverse.Metrics != roundTrip.Generated.Receipt.Metrics {
		t.Fatal("reverse metrics are not bound to generated receipt metrics")
	}
	if err := ValidateSemanticReverseObservation(roundTrip.Reverse); err != nil {
		t.Fatal(err)
	}

	mutatedSource := append(append([]byte(nil), source...), []byte("\n// authority mutation\n")...)
	if _, err := ReverseObserveSemanticOperationEnvelope(mutatedSource, outputDir); err == nil {
		t.Fatal("expected mutated authority to be rejected")
	}
	if err := os.WriteFile(filepath.Join(outputDir, "operation-receipt.json"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReverseObserveSemanticOperationEnvelope(source, outputDir); err == nil {
		t.Fatal("expected tampered artifact to be rejected")
	}
}
