package proposal

import (
	"reflect"

	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

type strategyFacts struct {
	Decision, PlanDigest, ReceiptDigest string
	Writes, Promotion                   bool
}

func strategyCoordinates(repository, subject string, first, replay strategy.Plan, receipt strategyverify.Receipt) ([]Coordinate, strategyFacts, error) {
	exact := first.Schema == strategy.PlanSchema && first.Repository == repository && first.SubjectSHA == subject && first.Input.SourceMetricsDigest != "" && first.Input.InterventionDigest != ""
	verified := validReceipt(receipt) && receipt.Schema == strategyverify.ReceiptSchema && receipt.Status == "VERIFIED" && receipt.PlanDigest == first.Digest && receipt.SourceMetricsDigest == first.Input.SourceMetricsDigest && receipt.InterventionDigest == first.Input.InterventionDigest && receipt.BindingCount == len(first.Bindings) && receipt.CandidateCount == len(first.Candidates) && receipt.SelectedProofChoice == first.Selection.ProofChoice
	replayed := strategy.ValidPlanDigest(first) && strategy.ValidPlanDigest(replay) && reflect.DeepEqual(first, replay)
	trilemma, unresolved := validTrilemma(first)
	values := []struct {
		ok, unknown bool
		reason      string
		evidence    any
	}{
		{exact, false, "EXACT_STRATEGY_SUBJECT_BOUND", []any{first.Schema, first.Repository, first.SubjectSHA, first.Input}},
		{verified, false, "STRATEGY_RECEIPT_VERIFIED", receipt},
		{replayed, false, "STRATEGY_REPLAY_EQUAL", []string{first.Digest, replay.Digest}},
		{trilemma, unresolved, "CONCEPT_TRILEMMA_BOUND", []any{first.Policy, first.Candidates, first.Selection}},
	}
	result := make([]Coordinate, 0, len(values))
	for index, value := range values {
		status, reason := coordinateStatus(value.ok, value.unknown, value.reason)
		coordinate, err := makeCoordinate(index, status, reason, value.evidence)
		if err != nil {
			return nil, strategyFacts{}, err
		}
		result = append(result, coordinate)
	}
	facts := strategyFacts{first.Selection.Decision, first.Digest, receipt.Digest, first.RepositoryWorkspaceWrites || receipt.RepositoryWorkspaceWrites, first.PromotionAuthorized || receipt.PromotionAuthorized}
	return result, facts, nil
}
