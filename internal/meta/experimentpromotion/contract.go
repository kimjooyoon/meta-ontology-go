package experimentpromotion

import "fmt"

var requiredReceiptFields = []string{
	"pr_number", "head_sha", "source_raw_digest", "source_semantic_digest",
	"producer_id", "consumer_package", "consumer_imports", "claim_transition_digest",
	"actions.run_url", "actions.job_url", "actions.conclusion",
	"artifact.bytes", "artifact.path", "artifact.digest",
}

var notClaimed = []string{
	"overall promotion score",
	"weighted gate score",
	"improvement rate",
	"aggregate estimate",
	"PR title or body as evidence",
	"network conclusion cache",
}

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.Version != 1 || contract.SourcePath != SourcePath {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_IDENTITY_INVALID")
	}
	if !sameStrings(contract.Experiments, experimentIDs()) || !sameStrings(contract.Gates, GateIDs) {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_VOCABULARY_INVALID")
	}
	if contract.ExperimentDenominator != ExperimentCount || contract.GateSlotDenominator != GateSlotCount {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_DENOMINATOR_INVALID")
	}
	if !sameStrings(contract.RequiredReceiptFields, requiredReceiptFields) {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_RECEIPT_FIELDS_INVALID")
	}
	if !sameStrings(contract.NotClaimed, notClaimed) {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_GUARDRAILS_INVALID")
	}
	return nil
}
