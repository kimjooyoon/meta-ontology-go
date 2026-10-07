package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// AssemblyCounts uses whole construction cases, not individual record fields.
// Budget is the smaller of the declared attempt budget and available candidates.
type AssemblyCounts struct {
	Matched int `json:"matched"`
	Total   int `json:"total"`
	Best    int `json:"best"`
	Scored  int `json:"scored"`
	Budget  int `json:"budget"`
}

type RecordConstructionObservation struct {
	Activity       string         `json:"activity"`
	ActivityID     string         `json:"activity_id"`
	AttemptIndex   int            `json:"attempt_index"`
	CandidateMask  uint16         `json:"candidate_mask"`
	Selected       bool           `json:"selected"`
	AttemptStatus  string         `json:"attempt_status,omitempty"`
	DeclaredBudget int            `json:"declared_budget"`
	CandidateCount int            `json:"candidate_count"`
	Counts         AssemblyCounts `json:"counts"`
}

// ObserveRecordConstruction reconstructs candidate scores and the selected
// projection before exposing counts to a Gooo tool. It performs no inference
// and makes no claim about historical native runtime observations.
func ObserveRecordConstruction(ctx context.Context, source []byte, prior Composition) ([]RecordConstructionObservation, error) {
	if ctx == nil {
		return nil, fmt.Errorf("construction observation requires a context")
	}
	if err := VerifyComposition(ctx, "workspace.gooo", source, prior); err != nil {
		return nil, err
	}
	var rows []RecordConstructionObservation
	for _, step := range prior.Steps {
		report := step.Generation.Report
		spec, err := bodycodegen.SourceAssembly(ctx, "workspace.gooo", source, report.Activity)
		if err != nil {
			return nil, err
		}
		if spec == nil {
			continue
		}
		r := report.RecordAssembly
		if r == nil {
			return nil, fmt.Errorf("construction observation currently requires record choices; activity %s uses another profile", report.Activity)
		}
		best := 0
		for i, attempt := range r.Attempts {
			best = max(best, attempt.Passed)
			counts := AssemblyCounts{Matched: attempt.Passed, Total: r.Total, Best: best,
				Scored: i + 1, Budget: min(spec.MaxAttempts, len(r.Ranking))}
			rows = append(rows, RecordConstructionObservation{Activity: report.Activity, ActivityID: report.ActivityID,
				AttemptIndex: i, CandidateMask: attempt.Mask, Selected: attempt.Mask == r.SelectedMask,
				AttemptStatus: attempt.Status, DeclaredBudget: spec.MaxAttempts, CandidateCount: len(r.Ranking), Counts: counts})
		}
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("saved composition has no record construction observations")
	}
	return rows, nil
}
