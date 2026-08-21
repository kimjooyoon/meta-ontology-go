package generation

import "fmt"

func candidateKey(candidate candidate) string {
	return fmt.Sprintf("%020d\x00%s", candidate.binding.Priority, indicatorKey(candidate.indicator))
}

func actionFor(candidate candidate, id string) Action {
	return Action{IndicatorID: id, MetricID: candidate.indicator.MetricID, Subject: candidate.indicator.Subject,
		Operation: candidate.binding.Operation, IndependenceGroupID: candidate.binding.IndependenceGroupID,
		ProofChoice: candidate.binding.ProofChoice, Executor: candidate.binding.Executor, Evaluator: candidate.binding.Evaluator,
		RequiredIndicatorIDs: append([]string{}, candidate.binding.RequiredIndicatorIDs...), ReceiptRequired: candidate.binding.ReceiptRequired,
		Priority: candidate.binding.Priority}
}
