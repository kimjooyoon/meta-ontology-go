package ciplanusecase

import "fmt"

func buildProofs(summary Summary, contract Contract) []Proof {
	foundation := status(contract.Schema == ContractSchema && contract.Denominator == 12 && summary.ResourceSamples == 12)
	coherence := status(summary.CasesSatisfied == 12 && summary.DeterministicReplays == 12 && summary.GoldenPlans == 4 && summary.GeneratedReplays == 1)
	regression := status(summary.PassDecisions == 4 && summary.FailClosedDecisions == 4 && summary.UnknownDecisions == 4 && summary.RepositoryWrites == 0 && summary.MutationAuthority == 0)
	return []Proof{
		{Choice: "FOUNDATION", Status: foundation, Evidence: []string{fmt.Sprintf("contract=%d resource-samples=%d", contract.Denominator, summary.ResourceSamples)}},
		{Choice: "COHERENCE", Status: coherence, Evidence: []string{fmt.Sprintf("cases=%d replays=%d golden=%d generated=%d", summary.CasesSatisfied, summary.DeterministicReplays, summary.GoldenPlans, summary.GeneratedReplays)}},
		{Choice: "REGRESSION", Status: regression, Evidence: []string{fmt.Sprintf("pass=%d fail=%d unknown=%d writes=%d authority=%d", summary.PassDecisions, summary.FailClosedDecisions, summary.UnknownDecisions, summary.RepositoryWrites, summary.MutationAuthority)}},
	}
}

func status(value bool) string {
	if value {
		return "SATISFIED"
	}
	return "UNSATISFIED"
}
