package generation

import "fmt"

func planIndicatorDecisionLedgerProvenance(plan Plan) (string, int) {
	if plan.IndicatorDecisionLedger == nil {
		return "", 0
	}
	return plan.IndicatorDecisionLedger.Digest, plan.IndicatorDecisionLedger.IndicatorCount
}

func validateExecutionIndicatorLedgerProvenance(manifest ExecutionManifest) error {
	required := manifest.Decision == ExecutionDecisionFixedPoint || manifest.Decision == ExecutionDecisionProposed
	if required {
		if !validIndicatorDecisionLedgerDigest(manifest.IndicatorDecisionLedgerDigest) || manifest.IndicatorDecisionLedgerCount < 0 {
			return fmt.Errorf("executable manifest has invalid indicator decision ledger provenance")
		}
		return nil
	}
	if manifest.IndicatorDecisionLedgerDigest != "" || manifest.IndicatorDecisionLedgerCount != 0 {
		return fmt.Errorf("non-executable manifest carries indicator decision ledger provenance")
	}
	return nil
}
