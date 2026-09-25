package predecessorresolution

import "fmt"

func validateInput(input Input) error {
	if input.Repository == "" || !validSHA(input.CurrentHeadSHA) ||
		!validSHA(input.ImmediatePredecessorSHA) ||
		input.CurrentHeadSHA == input.ImmediatePredecessorSHA ||
		input.SearchLimit != SearchLimit || len(input.Attempts) == 0 ||
		len(input.Attempts) > SearchLimit {
		return fmt.Errorf("ancestor resolution identity malformed")
	}
	return nil
}

func validateAttempt(input Input, index int, attempt Attempt) (attemptKind, error) {
	if attempt.Depth != index || !validSHA(attempt.AncestorSHA) ||
		(index == 0 && attempt.AncestorSHA != input.ImmediatePredecessorSHA) {
		return attemptBlocked, fmt.Errorf("ancestor attempt order malformed")
	}
	kind, err := classifyAttempt(input, attempt)
	if err != nil {
		return kind, err
	}
	last := index == len(input.Attempts)-1
	if kind == attemptMissing && !last && !validSHA(attempt.ParentSHA) {
		return kind, fmt.Errorf("missing ancestor parent malformed")
	}
	if kind != attemptMissing && !last {
		return kind, fmt.Errorf("resolution continued after terminal evidence")
	}
	if kind != attemptMissing && attempt.ParentSHA != "" {
		return kind, fmt.Errorf("terminal ancestor has parent continuation")
	}
	if kind == attemptSelected && !last {
		return kind, fmt.Errorf("resolution continued after selection")
	}
	return kind, nil
}

func addAttemptSummary(summary *Summary, attempt Attempt, kind attemptKind) {
	selectionSummary := attempt.Selection.Report.Summary
	summary.AmbiguousCandidates += selectionSummary.AmbiguousCandidates
	summary.RepositoryWrites += selectionSummary.RepositoryWrites
	if kind == attemptMissing {
		summary.MissingAttempts++
	}
	if kind == attemptSelected {
		summary.ValidCandidates += selectionSummary.ValidCandidates
	}
}
