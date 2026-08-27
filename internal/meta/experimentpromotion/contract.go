package experimentpromotion

import "fmt"

var requiredReceiptFields = []string{
	"pr_number", "claim_address", "evidence_class", "head_sha", "source_raw_digest", "source_semantic_digest",
	"producer_id", "consumer_package", "consumer_imports", "claim_class", "claim_transition_digest",
	"procedure_id", "procedure_source_path", "procedure_source_bytes", "procedure_source_digest",
	"procedure_algorithm_id", "procedure_algorithm_digest", "target_address",
	"actions.repository", "actions.pr_number", "actions.head_sha", "actions.workflow_id", "actions.workflow_name",
	"actions.run_id", "actions.job_id", "actions.run_url", "actions.job_url", "actions.conclusion",
	"actions.artifact_id", "actions.artifact_name", "actions.artifact_digest", "actions.raw", "actions.raw_digest",
	"artifact.bytes", "artifact.path", "artifact.digest", "artifact.target_address", "artifact.artifact_id", "artifact.artifact_name", "artifact.raw",
}

var notClaimed = []string{
	"overall promotion score", "weighted gate score", "improvement rate", "aggregate estimate",
	"PR title or body as evidence", "network conclusion cache", "fixture as current evidence",
}

// ExpectedExperiments is used only to validate the checked-in contract. The
// producer's authority remains parseSource(main.gooo), which reconstructs the
// actual identity records before this expectation is compared.
func ExpectedExperiments() []ExperimentIdentity {
	topics := []string{
		"claim-lifecycle", "resolution-descent", "provenance", "capability", "phase-separation", "hygiene", "proof-choice", "counterexample-first", "causal-ci", "ambiguity", "freshness", "reproducibility", "observer", "partial-knowledge", "proof-artifact", "meta-circular", "reflective-sandbox", "policy", "invariant-transform", "semantic-delta", "audience", "claim-dependency", "refutation", "denominator", "external-oracle", "fixed-point", "resource", "quorum", "causal-explanation", "portfolio",
	}
	prs := []int{549, 548, 545, 555, 544, 543, 546, 559, 550, 552, 564, 547, 558, 542, 569, 567, 551, 560, 553, 562, 563, 566, 554, 570, 541, 556, 561, 568, 565, 557}
	result := make([]ExperimentIdentity, ExperimentCount)
	for index := range result {
		result[index] = ExperimentIdentity{
			ID: fmt.Sprintf("experiment-%02d", index+1), PRNumber: prs[index], Topic: topics[index],
			ClaimAddress: fmt.Sprintf("github://kimjooyoon/meta-ontology-go/pull/%d#%s", prs[index], topics[index]),
		}
	}
	return result
}

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.Version != 2 || contract.SourcePath != SourcePath {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_IDENTITY_INVALID")
	}
	if !sameIdentities(contract.Experiments, ExpectedExperiments()) || !sameStrings(contract.Gates, GateIDs) {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_VOCABULARY_INVALID")
	}
	if contract.ExperimentDenominator != ExperimentCount || contract.GateSlotDenominator != GateSlotCount {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_DENOMINATOR_INVALID")
	}
	if !sameStrings(contract.RequiredReceiptFields, requiredReceiptFields) || !sameStrings(contract.NotClaimed, notClaimed) {
		return fmt.Errorf("EXPERIMENT_PROMOTION_CONTRACT_RECEIPT_FIELDS_INVALID")
	}
	return nil
}

func contractMatchesSource(contract Contract, source SourceProjection) bool {
	return sameIdentities(contract.Experiments, source.Experiments) && sameStrings(contract.Gates, source.Gates) && contract.ExperimentDenominator == len(source.Experiments) && contract.GateSlotDenominator == len(source.Experiments)*len(source.Gates)
}

func sameIdentities(left, right []ExperimentIdentity) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
