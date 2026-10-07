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
	f := report.BodyFill
	var rows []ConstructionObservation
	best := 0
	for _, c := range f.CandidateScores {
		best = max(best, c.TestCasesPassed)
	}
	for i, c := range f.CandidateScores {
		rows = append(rows, ConstructionObservation{Activity: report.Activity, ActivityID: report.ActivityID,
			Profile: "source_fill", View: "scored_set", CandidateIndex: &i, CandidateID: c.ID,
			Selected: c.ID == f.SelectedCandidateID, Proposed: c.ID == f.ProposedCandidateID, ScoringCompleted: true,
			DeclaredBudget: len(f.CandidateScores), CandidateCount: len(f.CandidateScores),
			Counts: AssemblyCounts{Matched: c.TestCasesPassed, Total: c.TestCasesTotal, Best: best,
				Scored: len(f.CandidateScores), Budget: len(f.CandidateScores)}})
	}
	return rows
}
