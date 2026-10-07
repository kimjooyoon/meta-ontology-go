package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// AssemblyCounts uses whole construction cases, not individual record fields.
// Budget caps attempts by available candidates and subtracts unscored attempts
// observed so far. Counts are meaningful only when ScoringCompleted is true.
type AssemblyCounts struct {
	Matched int `json:"matched"`
	Total   int `json:"total"`
	Best    int `json:"best"`
	Scored  int `json:"scored"`
	Budget  int `json:"budget"`
}

type ConstructionObservation struct {
	Activity         string         `json:"activity"`
	ActivityID       string         `json:"activity_id"`
	Profile          string         `json:"profile"`
	View             string         `json:"view"`
	AttemptIndex     *int           `json:"attempt_index,omitempty"`
	CandidateIndex   *int           `json:"candidate_index,omitempty"`
	CandidateMask    *uint16        `json:"candidate_mask,omitempty"`
	CandidateID      string         `json:"candidate_id,omitempty"`
	Selected         bool           `json:"selected"`
	Proposed         bool           `json:"proposed"`
	ScoringCompleted bool           `json:"scoring_completed"`
	InputIndex       *int           `json:"input_index,omitempty"`
	AttemptStatus    string         `json:"attempt_status,omitempty"`
	Reason           string         `json:"reason,omitempty"`
	DeclaredBudget   int            `json:"declared_budget"`
	CandidateCount   int            `json:"candidate_count"`
	Counts           AssemblyCounts `json:"counts"`
}

type RecordConstructionObservation = ConstructionObservation

// ObserveRecordConstruction reconstructs candidate scores and the selected
// projection before exposing counts to a Gooo tool. It performs no inference
// and makes no claim about historical native runtime observations.
func ObserveRecordConstruction(ctx context.Context, source []byte, prior Composition) ([]RecordConstructionObservation, error) {
	rows, err := ObserveConstruction(ctx, source, prior)
	if err == nil && len(rows) == 0 {
		err = fmt.Errorf("saved composition has no construction observations")
	}
	return rows, err
}

// ObserveConstruction verifies the whole composition before exposing source
// construction observations. A plain composition returns no observations.
func ObserveConstruction(ctx context.Context, source []byte, prior Composition) ([]ConstructionObservation, error) {
	if ctx == nil {
		return nil, fmt.Errorf("construction observation requires a context")
	}
	if err := VerifyComposition(ctx, "workspace.gooo", source, prior); err != nil {
		return nil, err
	}
	var rows []ConstructionObservation
	for _, step := range prior.ConstructionSteps() {
		report := step.Generation.Report
		if report.RecordAssembly == nil && report.BodySearch == nil && report.BodyFill == nil {
			continue
		}
		spec, err := bodycodegen.SourceAssembly(ctx, "workspace.gooo", source, report.Activity)
		if err != nil {
			return nil, err
		}
		if spec == nil {
			continue
		}
		switch {
		case report.RecordAssembly != nil:
			rows = append(rows, observeRecordAttempts(report, spec.MaxAttempts)...)
		case report.BodySearch != nil:
			rows = append(rows, observeSearchAttempts(report, spec.MaxAttempts)...)
		case report.BodyFill != nil:
			rows = append(rows, ObserveVerifiedFill(report)...)
		default:
			return nil, fmt.Errorf("construction observation does not support activity %s's assembly profile", report.Activity)
		}
	}
	return rows, nil
}

func observeRecordAttempts(report bodycodegen.Report, budget int) []ConstructionObservation {
	r := report.RecordAssembly
	var rows []ConstructionObservation
	best, scored, rejected := 0, 0, 0
	for i, a := range r.Attempts {
		row := ConstructionObservation{Activity: report.Activity, ActivityID: report.ActivityID,
			Profile: "record_choices", View: "attempt_prefix", AttemptIndex: &i, CandidateMask: &a.Mask,
			Proposed: r.Prediction != nil && a.Mask == r.Prediction.Mask,
			Selected: a.Mask == r.SelectedMask, AttemptStatus: a.Status, Reason: a.Reason,
			DeclaredBudget: budget, CandidateCount: len(r.Ranking), ScoringCompleted: a.Total > 0}
		if row.ScoringCompleted {
			scored++
			best = max(best, a.Passed)
			row.Counts = AssemblyCounts{Matched: a.Passed, Total: a.Total, Best: best,
				Scored: scored, Budget: min(budget, len(r.Ranking)) - rejected}
		} else {
			rejected++
		}
		rows = append(rows, row)
	}
	return rows
}
