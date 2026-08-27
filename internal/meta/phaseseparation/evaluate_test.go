package phaseseparation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPreservesPhaseClaimsAndCatchesLeaks(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	sourceBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "main.gooo"))
	leakBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "leaks.gooo"))
	report := Build("main.gooo", sourceBytes, "leaks.gooo", leakBytes, "0123456789012345678901234567890123456789")
	if report.Decision != DecisionPass || report.Summary.CleanCasesPassed != ExpectedCleanCases || report.Summary.LeakageCasesCaught != ExpectedLeakageCases {
		t.Fatalf("phase witness = %#v", report)
	}
	if report.Summary.ClaimTransitionsPreserved != ExpectedClaimTransitions || len(report.Transitions) != ExpectedClaimTransitions {
		t.Fatalf("claim transitions = %#v", report.Transitions)
	}
	for _, transition := range report.Transitions {
		if transition.MetaOperation != report.MetaOperation || transition.ProofChoice != report.ProofChoice || !transition.Preserved {
			t.Fatalf("unbound claim transition = %#v", transition)
		}
	}
}

func TestBuildKeepsMalformedSourceUnknown(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	unknownBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "unknown.gooo"))
	leakBytes := mustRead(t, filepath.Join(root, "examples", "phase-separation-witness", "leaks.gooo"))
	report := Build("unknown.gooo", unknownBytes, "leaks.gooo", leakBytes, "0123456789012345678901234567890123456789")
	if report.Decision != DecisionUnknown || report.Coordinate != (Coordinate{"SOURCE", "PARSE", ReasonUnknownSource}) {
		t.Fatalf("unknown report = %#v", report)
	}
	if report.Summary.RepositoryWrites != 0 || report.Authority != (Authority{}) {
		t.Fatalf("unknown authority = %#v", report.Authority)
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
