package transformationeffect

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayDifferenceDistinguishesMissingKeyFromNull(t *testing.T) {
	path, expected, observed := firstReplayDifference(
		"$", map[string]any{}, map[string]any{"field": nil},
	)
	if path != "$.field" || expected != "<missing>" || observed != "null" {
		t.Fatalf("difference = %q, %q, %q", path, expected, observed)
	}
}

func TestReplayDiagnosticPreservesTypedFailureAxes(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "ledger.json")
	if err := WriteReplayDiagnostic(output, errors.New("uncataloged")); err != nil {
		t.Fatal(err)
	}
	unknown := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if unknown.Decision != "UNKNOWN" || unknown.Resolution != "LOWER_RESOLUTION" ||
		unknown.UnknownClass != "UNCATALOGED_CAUSE" || unknown.NextOperation != "report-counterexample" ||
		unknown.BlockedBy == nil || len(unknown.BlockedBy) != 0 {
		t.Fatalf("unknown diagnostic = %#v", unknown)
	}

	divergence := &replayDivergence{Stage: "validate-inputs", Step: "compare-receipts", Path: "$.failures", Expected: "[]", Observed: "null"}
	if err := WriteReplayDiagnostic(output, divergence); err != nil {
		t.Fatal(err)
	}
	refuted := readReplayDiagnostic(t, filepath.Join(directory, "replay-diagnostic.json"))
	if refuted.Decision != "REFUTED" || refuted.Resolution != "EXACT" ||
		refuted.FieldPath != "$.failures" || refuted.UnknownClass != "" {
		t.Fatalf("refuted diagnostic = %#v", refuted)
	}
}

func readReplayDiagnostic(t *testing.T, path string) ReplayDiagnostic {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostic ReplayDiagnostic
	if err := json.Unmarshal(payload, &diagnostic); err != nil {
		t.Fatal(err)
	}
	return diagnostic
}
