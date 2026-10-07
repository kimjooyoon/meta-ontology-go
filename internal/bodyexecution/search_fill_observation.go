package bodyexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"

func observeSearchAttempts(report bodycodegen.Report, budget int) []ConstructionObservation {
	s := report.BodySearch
	var rows []ConstructionObservation
	best, scored, rejected := 0, 0, 0
	for i, a := range s.Attempts {
		row := ConstructionObservation{Activity: report.Activity, ActivityID: report.ActivityID,
			Profile: "ir_search", View: "attempt_prefix", AttemptIndex: &i, CandidateID: a.CandidateID,
			Selected: a.CandidateID == s.SelectedCandidateID, ScoringCompleted: a.ScoringCompleted,
			Reason: a.Error, DeclaredBudget: budget, CandidateCount: s.CandidateCount}
		if a.ScoringCompleted {
			scored++
			best = max(best, a.TestCasesPassed)
			row.Counts = AssemblyCounts{Matched: a.TestCasesPassed, Total: a.TestCasesTotal, Best: best,
				Scored: scored, Budget: min(budget, s.CandidateCount) - rejected}
		} else {
			rejected++
		}
		rows = append(rows, row)
	}
	return rows
}

// ObserveVerifiedFill must follow source-fill replay. All fill candidates were
// scored before selection, so this is a scored set, not a chronological search.
func ObserveVerifiedFill(report bodycodegen.Report) []ConstructionObservation {
	return observeVerifiedFill(report, "source_fill")
}

// ObserveVerifiedExternalFill must follow replay with the retained external plan.
// Its candidate scores use the same whole-case units as source-owned fills.
func ObserveVerifiedExternalFill(report bodycodegen.Report) []ConstructionObservation {
	return observeVerifiedFill(report, "external_fill")
}

func observeVerifiedFill(report bodycodegen.Report, profile string) []ConstructionObservation {
	f := report.BodyFill
	var rows []ConstructionObservation
	best := 0
	for _, c := range f.CandidateScores {
		best = max(best, c.TestCasesPassed)
	}
	for i, c := range f.CandidateScores {
		rows = append(rows, ConstructionObservation{Activity: report.Activity, ActivityID: report.ActivityID,
			Profile: profile, View: "scored_set", CandidateIndex: &i, CandidateID: c.ID,
			Selected: c.ID == f.SelectedCandidateID, Proposed: c.ID == f.ProposedCandidateID, ScoringCompleted: true,
			DeclaredBudget: len(f.CandidateScores), CandidateCount: len(f.CandidateScores),
			Counts: AssemblyCounts{Matched: c.TestCasesPassed, Total: c.TestCasesTotal, Best: best,
				Scored: len(f.CandidateScores), Budget: len(f.CandidateScores)}})
	}
	return rows
}
