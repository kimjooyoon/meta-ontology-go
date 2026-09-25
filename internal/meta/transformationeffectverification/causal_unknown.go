package transformationeffectverification

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"

func validateCausalUnknownProjection(plan generation.Plan, ledger ledger, report generation.ReceiptReport) error {
	projection, err := deriveCausalUnknownProjection(report)
	if err != nil {
		return bindingFailure("receipts.unknowns", "canonical causal projection", err.Error())
	}
	if ledger.DirectUnknownCount != projection.DirectUnknownCount ||
		ledger.DependencyBlockedUnknownCount != projection.DependencyBlockedUnknownCount ||
		ledger.UnknownCausalDigest != projection.Digest {
		return bindingFailure("ledger.unknown_causal_digest", projection.Digest, ledger.UnknownCausalDigest)
	}
	selected := make(map[string]map[string]bool, len(plan.Selected))
	for _, action := range plan.Selected {
		selected[action.IndicatorID] = make(map[string]bool, len(action.RequiredIndicatorIDs))
		for _, required := range action.RequiredIndicatorIDs {
			selected[action.IndicatorID][required] = true
		}
	}
	for _, record := range projection.Records {
		required, ok := selected[record.ActionIndicatorID]
		if !ok || !required[record.RequiredIndicatorID] {
			return bindingFailure("receipts.unknowns", "selected required obligation", causalUnknownKey(record))
		}
	}
	if ledger.OperationOutcome == "MIXED_CLOSED_REFUTED" &&
		(projection.DirectUnknownCount != 0 || projection.DependencyBlockedUnknownCount != len(report.Unknowns)) {
		return bindingFailure("ledger.unknown_causal_counts", "all unknowns dependency-blocked", "direct or unbound unknowns present")
	}
	return nil
}
