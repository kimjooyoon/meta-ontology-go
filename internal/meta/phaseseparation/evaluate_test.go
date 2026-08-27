package phaseseparation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildReconstructsRealGoooCorpus(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	mainBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "main.gooo"))
	leakBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "leaks.gooo"))
	unknownBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "unknown.gooo"))
	report := Build("main.gooo", mainBytes, "leaks.gooo", leakBytes, "unknown.gooo", unknownBytes, "0123456789012345678901234567890123456789", CISnapshot{})
	if report.Decision != DecisionPass || report.Summary.SourceCasesProcessed != ExpectedSourceCases || report.Summary.LeakageRejections != ExpectedLeakageCases {
		t.Fatalf("phase witness = %#v", report)
	}
	if report.Summary.ClaimTransitionsPreserved != ExpectedClaimTransitions || report.Summary.ExplicitClaimTransfers != ExpectedClaimTransitions {
		t.Fatalf("claim transitions = %#v", report.Transitions)
	}
	if report.Summary.SemanticCausality != 1 || report.Summary.NonsemanticPreservation != 1 {
		t.Fatalf("interventions = %#v / %#v", report.SemanticIntervention, report.NonsemanticIntervention)
	}
}

func TestBuildKeepsUnknownProbeOpen(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	mainBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "main.gooo"))
	leakBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "leaks.gooo"))
	unknownBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "unknown.gooo"))
	report := Build("main.gooo", mainBytes, "leaks.gooo", leakBytes, "unknown.gooo", unknownBytes, "0123456789012345678901234567890123456789", CISnapshot{})
	if report.Unknown.Decision != DecisionUnknown || report.Unknown.Resolution != ResolutionLower || report.Unknown.Coordinate != (Coordinate{"SOURCE", "PARSE", ReasonUnknownProvenance}) || report.Unknown.ClaimState != StateOpen {
		t.Fatalf("unknown report = %#v", report.Unknown)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
