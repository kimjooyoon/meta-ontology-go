package bodycodegen

import (
	"context"
	"errors"

	"github.com/kimjooyoon/gooo-decision-runtime/bodyplan"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

// This adapter preserves the compiler's partial-result and total-budget rules.
// ContractSession owns one ranking; advancing it never calls the model again.
func searchContractPaths(ctx context.Context, prepared *pathplan.PreparedPlan, model *conditionPathModel,
	cases []pathplan.TestCase, total, step int) (pathplan.SearchResult, *bodyplan.Program,
	*pathplan.ContractRanking, []pathplan.ContractProgress, error) {
	session, err := model.newContractSession(ctx, prepared, cases)
	if err != nil {
		return pathplan.SearchResult{}, nil, nil, nil, err
	}
	ranking := session.Ranking()
	initial, err := session.Observe()
	if err != nil {
		return pathplan.SearchResult{}, nil, &ranking, nil, err
	}
	records := []pathplan.ContractProgress{initial}
	result := contractSearchResult(initial.SessionProgress, nil)
	var body *bodyplan.Program
	var attempts []pathplan.SearchAttempt
	for len(attempts) < total {
		progress, selected, failure := session.Advance(ctx, min(step, total-len(attempts)))
		if progress.Schema != "" {
			attempts = append(attempts, progress.NewAttempts...)
			records = append(records, progress)
			result, body = contractSearchResult(progress.SessionProgress, attempts), selected
		}
		if failure != nil && !errors.Is(failure, pathplan.ErrNoTypedCandidate) {
			return result, body, &ranking, records, failure
		}
		if progress.Status == "TRAINING_COMPLETE" || progress.Exhausted || len(attempts) >= total {
			return result, body, &ranking, records, failure
		}
		if len(progress.NewAttempts) == 0 {
			return result, body, &ranking, records, errors.New("contract session made no candidate progress")
		}
	}
	return result, body, &ranking, records, nil
}

func contractSearchResult(p pathplan.SessionProgress, attempts []pathplan.SearchAttempt) pathplan.SearchResult {
	return pathplan.SearchResult{Schema: "gooo/typed-path-tdd-search/v1", Status: p.Status, Selection: p.Selection,
		DeclaredCombinations: p.Declared, Unattempted: p.Unattempted, Evaluated: p.Evaluated,
		TypeRejected: p.TypeRejected, ConditionRejected: p.ConditionRejected, SelectedTrainingPassed: p.SelectedPassed,
		TrainingTotal: p.Cases, Attempts: attempts, ModelAbstentionsObserved: p.ModelAbstentions,
		EligibleProbabilities: p.EligibleProbabilities, InitialProposals: p.InitialProposals}
}
