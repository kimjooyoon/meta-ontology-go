package languagesourceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

func evaluateCases(input Input) ([]CaseResult, Summary) {
	raws := [][]byte{input.Positive, input.Replay, input.UnknownEntry, input.InvalidSyntax}
	receipts := make([]sourceexecution.Receipt, len(raws))
	errors := make([]error, len(raws))
	for index, raw := range raws {
		receipts[index], errors[index] = decodeReceipt(raw)
	}
	results := []CaseResult{
		receiptCase(input.Contract.Cases[0], raws[0], receipts[0], errors[0]),
		replayCase(input.Contract.Cases[1], raws[0], raws[1], errors[0], errors[1]),
		receiptCase(input.Contract.Cases[2], raws[2], receipts[2], errors[2]),
		receiptCase(input.Contract.Cases[3], raws[3], receipts[3], errors[3]),
	}
	return results, summarizeCases(results, receipts)
}

func summarizeCases(results []CaseResult, receipts []sourceexecution.Receipt) Summary {
	summary := Summary{CasesTotal: len(results)}
	for _, result := range results {
		switch result.Status {
		case "SATISFIED":
			summary.CasesSatisfied++
		case "UNKNOWN":
			summary.Unknowns++
		default:
			summary.NotSatisfied++
		}
	}
	if results[0].Status == "SATISFIED" {
		summary.SourceExecutions, summary.ExecutionEvents = 1, len(receipts[0].Events)
	}
	if results[1].Status == "SATISFIED" {
		summary.DeterministicReplays = 1
	}
	for _, index := range []int{2, 3} {
		if results[index].Status == "SATISFIED" {
			summary.DiagnosticRejections++
		}
	}
	for _, receipt := range receipts {
		summary.RepositoryWrites += receipt.Effects.RepositoryWrites
		if receipt.Effects.MutationAuthority {
			summary.MutationAuthorities++
		}
	}
	return summary
}
