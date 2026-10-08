package bodyexecution

import (
	"context"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordConstructionObservationScoresUnusedLocalCandidates(t *testing.T) {
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
	if len(rows) != 7 {
		t.Fatal("fixture did not retain all seven evaluated candidates", rows)
	}
	for i, row := range rows {
		if !row.ScoringCompleted || row.AttemptStatus != "" || row.Reason != "" ||
			row.Counts.Scored != i+1 || row.Counts.Budget != 8 || row.Counts.Total != 5 {
			t.Fatal("valid unused-local candidate lost its finite observation", row)
		}
	}
	if rows[3].CandidateMask == nil || *rows[3].CandidateMask != 3 || rows[3].Counts.Matched != 2 ||
		!rows[6].Selected || rows[6].Counts.Matched != 5 {
		t.Fatal("partial and complete case results changed", rows)
	}
	// A forged rejection cannot replace the replayed result of this valid candidate.
	prior.Steps[0].Generation.Report.RecordAssembly.Attempts[3] = bodycodegen.RecordAssemblyAttempt{
		Mask: 3, Status: "TYPECHECK_FAILED", Reason: "unused local",
	}
	if _, err = ObserveConstruction(context.Background(), source, prior); err == nil {
		t.Fatal("forged type rejection became a construction observation")
	}
}

func TestRecordConstructionObservationRetainsTypeRejectionsWithoutScores(t *testing.T) {
	// Synthetic attempts isolate the observation adapter. Actual type rejection
	// and continuation are exercised in bodycodegen's candidate tests.
	rejected := bodycodegen.RecordAssemblyAttempt{Status: "TYPECHECK_FAILED", Reason: "field type mismatch"}
	zero := bodycodegen.RecordAssemblyAttempt{Total: 5}
	partial := bodycodegen.RecordAssemblyAttempt{Passed: 2, Total: 5}
	for _, tc := range []struct {
		name     string
		attempts []bodycodegen.RecordAssemblyAttempt
	}{
		{"before_scores", []bodycodegen.RecordAssemblyAttempt{rejected, zero, partial}},
		{"between_scores", []bodycodegen.RecordAssemblyAttempt{zero, rejected, partial}},
		{"after_scores", []bodycodegen.RecordAssemblyAttempt{zero, partial, rejected}},
		{"all_rejected", []bodycodegen.RecordAssemblyAttempt{rejected, rejected}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, budget := range []int{3, 8} {
				report := bodycodegen.Report{RecordAssembly: &bodycodegen.RecordAssemblyReceipt{
					Ranking: []uint16{0, 1, 2, 3}, Attempts: tc.attempts,
				}}
				scored, rejected, best := 0, 0, 0
				for i, row := range observeRecordAttempts(report, budget) {
					attempt := tc.attempts[i]
					if attempt.Total == 0 {
						rejected++
						if row.ScoringCompleted || row.AttemptStatus != "TYPECHECK_FAILED" || row.Reason == "" ||
							row.Counts != (AssemblyCounts{}) || row.InputIndex != nil {
							t.Fatal("type rejection became a measured zero", row)
						}
						continue
					}
					scored++
					best = max(best, attempt.Passed)
					want := AssemblyCounts{Matched: attempt.Passed, Total: 5, Best: best,
						Scored: scored, Budget: min(budget, 4) - rejected}
					if !row.ScoringCompleted || row.Counts != want {
						t.Fatal("scored count or remaining capacity included a rejected candidate", row, want)
					}
				}
			}
		})
	}
}
