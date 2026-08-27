package artifactcoverage

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateSelectsCanonicalGapDeterministically(t *testing.T) {
	action, observations := coverageFixture()
	report, err := Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FIXED_POINT" || report.Summary.CanonicalOperations != 6 ||
		report.Summary.CanonicalCoverageBasisPoints != 10000 {
		t.Fatalf("fixed point report = %#v", report)
	}
	first, _ := Marshal(report)
	replay, err := Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := Marshal(replay)
	if string(first) != string(second) {
		t.Fatal("coverage report replay differs")
	}
	filtered := observations.Artifacts[:0]
	for _, artifact := range observations.Artifacts {
		if !strings.HasPrefix(artifact.Name, "directory-kind-separation-") {
			filtered = append(filtered, artifact)
		}
	}
	observations.Artifacts = filtered
	report, err = Evaluate(filepath.Join("..", "..", ".."), action, observations)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "IMPROVE" || report.SelectedOperation != "separate-directory-kinds" {
		t.Fatalf("gap decision = %s selected = %s", report.Decision, report.SelectedOperation)
	}
}
