package governancesnapshot

import (
	"fmt"
	"sort"
)

var expectedContextNames = []string{
	"CI policy",
	"Semantic conformance",
	"gofmt",
	"go vet",
	"go test",
	"go test -race",
	"CI guardian",
}

func ExpectedCells() []CellSpec {
	return []CellSpec{
		{"OBSERVATION_SOURCE_PIN", "PinGovernanceObservationSource", "FOUNDATION", "DRIVER", "PinGovernanceObservationSource", "gooo://live-governance-snapshot/input/observation-source", "gooo://live-governance-snapshot/output/observation-source"},
		{"CONTRACT_MANIFEST_PIN", "PinGovernanceContractManifest", "FOUNDATION", "DRIVER", "PinGovernanceContractManifest", "gooo://live-governance-snapshot/input/contract-manifest", "gooo://live-governance-snapshot/output/contract-manifest"},
		{"DEV_BRANCH_PROTECTED", "ObserveDevBranchProtection", "FOUNDATION", "OUTCOME", "ObserveDevBranchProtection", "gooo://live-governance-snapshot/input/dev-branch", "gooo://live-governance-snapshot/output/dev-branch"},
		{"MAIN_BRANCH_PROTECTED", "ObserveMainBranchProtection", "FOUNDATION", "OUTCOME", "ObserveMainBranchProtection", "gooo://live-governance-snapshot/input/main-branch", "gooo://live-governance-snapshot/output/main-branch"},
		{"DEV_STATUS_ENFORCEMENT", "CompareDevStatusEnforcement", "COHERENCE", "GUARDRAIL", "CompareDevStatusEnforcement", "gooo://live-governance-snapshot/input/dev-status", "gooo://live-governance-snapshot/output/dev-status"},
		{"DEV_CONTEXT_SET", "CompareDevRequiredContexts", "COHERENCE", "OUTCOME", "CompareDevRequiredContexts", "gooo://live-governance-snapshot/input/dev-contexts", "gooo://live-governance-snapshot/output/dev-contexts"},
		{"MAIN_STATUS_ENFORCEMENT", "CompareMainStatusEnforcement", "COHERENCE", "GUARDRAIL", "CompareMainStatusEnforcement", "gooo://live-governance-snapshot/input/main-status", "gooo://live-governance-snapshot/output/main-status"},
		{"MAIN_CONTEXT_SET", "CompareMainRequiredContexts", "COHERENCE", "OUTCOME", "CompareMainRequiredContexts", "gooo://live-governance-snapshot/input/main-contexts", "gooo://live-governance-snapshot/output/main-contexts"},
		{"RULESET_INVENTORY", "ObserveRepositoryRulesets", "REGRESSION", "DRIVER", "ObserveRepositoryRulesets", "gooo://live-governance-snapshot/input/rulesets", "gooo://live-governance-snapshot/output/rulesets"},
		{"DISABLED_RULESET_AUTHORITY", "RejectDisabledRulesetAuthority", "REGRESSION", "GUARDRAIL", "RejectDisabledRulesetAuthority", "gooo://live-governance-snapshot/input/ruleset-authority", "gooo://live-governance-snapshot/output/ruleset-authority"},
		{"UNKNOWN_CAUSALITY", "PreserveGovernanceUnknownCausality", "REGRESSION", "GUARDRAIL", "PreserveGovernanceUnknownCausality", "gooo://live-governance-snapshot/input/unknown-causality", "gooo://live-governance-snapshot/output/unknown-causality"},
		{"HUMAN_DRIFT_REPORT", "PublishGovernanceDriftReport", "REGRESSION", "DRIVER", "PublishGovernanceDriftReport", "gooo://live-governance-snapshot/input/human-report", "gooo://live-governance-snapshot/output/human-report"},
	}
}

func ExpectedContexts() []string {
	return append([]string(nil), expectedContextNames...)
}

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.ID != "live-governance-snapshot-v1" || contract.GraphProgram == "" {
		return fmt.Errorf("governance snapshot contract identity is invalid")
	}
	if len(contract.Cells) != len(ExpectedCells()) {
		return fmt.Errorf("governance snapshot cell denominator is invalid")
	}
	for index, expected := range ExpectedCells() {
		if contract.Cells[index] != expected {
			return fmt.Errorf("governance snapshot cell %d is not canonical", index)
		}
	}
	if len(contract.Expected.StatusChecks) != 2 || len(contract.Expected.Rulesets) != 2 || contract.Expected.RequiredRulesetState != "disabled" {
		return fmt.Errorf("governance snapshot expectations are incomplete")
	}
	if contract.Expected.DefaultBranch != "dev" || contract.Expected.Repository == "" {
		return fmt.Errorf("governance snapshot default branch expectation is invalid")
	}
	for _, branch := range contract.Expected.StatusChecks {
		if branch.Enforcement != "everyone" || !sameStrings(branch.Contexts, expectedContextNames) || !branch.Protected {
			return fmt.Errorf("governance snapshot branch expectation is invalid")
		}
	}
	for _, ruleset := range contract.Expected.Rulesets {
		if ruleset.ID <= 0 || ruleset.Name != ruleset.Branch || ruleset.Enforcement != "disabled" {
			return fmt.Errorf("governance snapshot ruleset expectation is invalid")
		}
	}
	if len(contract.Source.Documentation) != 3 || len(contract.Source.APIVersions) != 2 || contract.Source.PayloadDigestModel != "canonical-json-v1" || len(contract.Source.Endpoints) != 4 {
		return fmt.Errorf("governance snapshot source authority is incomplete")
	}
	return nil
}

func sameStrings(left, right []string) bool {
	leftCopy, rightCopy := append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	if len(leftCopy) != len(rightCopy) {
		return false
	}
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}
