package provenance

import (
	"strings"
	"testing"
)

func validSPIFFEExecutionBoundaryInputPart01() SPIFFEExecutionBoundaryInputPart01 {
	return SPIFFEExecutionBoundaryInputPart01{
		WorkloadIdentity:          "spiffe://example.org/ns/meta/sa/compiler",
		WorkloadAttestationDigest: "sha256:" + strings.Repeat("a", 64),
		TrustBundleDigest:         "sha256:" + strings.Repeat("b", 64),
		NonAuthorizing:            true,
		GoooDeclarationDigest:     DigestSPIFFEGoooDeclarationPart01([]byte("entity Compiler")),
		IRDigest:                  "sha256:" + strings.Repeat("c", 64),
		GeneratedArtifactDigest:   "sha256:" + strings.Repeat("d", 64),
		ReverseObservationDigest:  "sha256:" + strings.Repeat("e", 64),
		ExecutionPlanDigest:       "sha256:" + strings.Repeat("f", 64),
		Metrics: SPIFFEExecutionBoundaryMetricsPart01{
			GeneratedArtifactCount:  1,
			ReverseObservationCount: 1,
			EvidenceStageCount:      8,
		},
	}
}

func TestGenerateAndReverseObserveSPIFFEExecutionBoundaryPreservesEvidence(t *testing.T) {
	envelope := GenerateSPIFFEExecutionBoundaryEnvelopePart01(validSPIFFEExecutionBoundaryInputPart01())
	if envelope.Status != SPIFFEExecutionBoundaryObservedPart01 || envelope.MissingStageIndex != -1 {
		t.Fatalf("valid envelope = %#v", envelope)
	}
	if err := envelope.Validate(); err != nil {
		t.Fatal(err)
	}
	reverse := ReverseObserveSPIFFEExecutionBoundaryEnvelopePart01(envelope)
	if reverse.Status != envelope.Status || reverse.Reason != envelope.Reason || reverse.GoooDeclarationDigest != envelope.GoooDeclarationDigest || reverse.IRDigest != envelope.IRDigest || reverse.ExecutionPlanDigest != envelope.ExecutionPlanDigest {
		t.Fatalf("reverse observation lost evidence: %#v", reverse)
	}
	if reverse.Metrics != envelope.Metrics || !reverse.NonAuthorizing {
		t.Fatalf("reverse observation lost metrics or authority boundary: %#v", reverse)
	}
}

func TestGenerateSPIFFEExecutionBoundaryKeepsFirstMissingStageUnknown(t *testing.T) {
	input := validSPIFFEExecutionBoundaryInputPart01()
	input.IRDigest = ""
	input.GeneratedArtifactDigest = ""
	envelope := GenerateSPIFFEExecutionBoundaryEnvelopePart01(input)
	if envelope.Status != SPIFFEExecutionBoundaryUnknownPart01 || envelope.MissingStageIndex != 4 || envelope.Reason != "SPIFFE_EXECUTION_BOUNDARY_IR_INVALID" {
		t.Fatalf("missing IR was promoted: %#v", envelope)
	}
	if err := envelope.Validate(); err != nil {
		t.Fatal(err)
	}
	reverse := ReverseObserveSPIFFEExecutionBoundaryEnvelopePart01(envelope)
	if reverse.Status != SPIFFEExecutionBoundaryUnknownPart01 || reverse.MissingStageIndex != 4 {
		t.Fatalf("reverse observation changed unknown stage: %#v", reverse)
	}
}

func TestReverseObserveSPIFFEExecutionBoundaryRejectsTampering(t *testing.T) {
	envelope := GenerateSPIFFEExecutionBoundaryEnvelopePart01(validSPIFFEExecutionBoundaryInputPart01())
	envelope.ExecutionPlanDigest = "sha256:" + strings.Repeat("0", 64)
	reverse := ReverseObserveSPIFFEExecutionBoundaryEnvelopePart01(envelope)
	if reverse.Status != SPIFFEExecutionBoundaryUnknownPart01 || reverse.Reason != "SPIFFE_EXECUTION_BOUNDARY_ENVELOPE_INVALID" {
		t.Fatalf("tampered envelope was observed: %#v", reverse)
	}
}

func TestGenerateSPIFFEExecutionBoundaryNeverAuthorizesIdentity(t *testing.T) {
	input := validSPIFFEExecutionBoundaryInputPart01()
	input.NonAuthorizing = false
	envelope := GenerateSPIFFEExecutionBoundaryEnvelopePart01(input)
	if envelope.Status != SPIFFEExecutionBoundaryUnknownPart01 || envelope.Reason != "SPIFFE_EXECUTION_BOUNDARY_AUTHORITY_INVALID" {
		t.Fatalf("authorizing input was accepted: %#v", envelope)
	}
}
