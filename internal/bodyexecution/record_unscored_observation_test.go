package bodyexecution

import (
	"context"
	"os"
	"testing"
)

func TestRecordConstructionObservationRetainsTypeRejectionsWithoutScores(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/record-candidate-continuation.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := os.ReadFile("../../examples/body-codegen/record-field-updates-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "workspace.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ObserveConstruction(context.Background(), source, prior)
	if err != nil {
		t.Fatal(err)
	}
	// Use the actual prefix to check how rejected attempts consume capacity.
	scored, rejected := 0, 0
	for _, row := range rows {
		if !row.ScoringCompleted {
			rejected++
			if row.AttemptStatus != "TYPECHECK_FAILED" || row.Reason == "" || row.Counts.Total != 0 || row.InputIndex != nil {
				t.Fatal("type rejection became a measured zero", row)
			}
			continue
		}
		scored++
		if row.Counts.Scored != scored || row.Counts.Budget != 8-rejected || row.Counts.Total != 5 {
			t.Fatal("scored count or remaining capacity included a rejected candidate", row)
		}
	}
	if rejected != 1 || scored != 6 {
		t.Fatal("fixture no longer reaches its complete finite result after a rejection", scored, rejected)
	}
}
