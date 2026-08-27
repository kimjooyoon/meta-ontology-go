package ambiguitybudget

import (
	"fmt"
	"sort"
)

func validateContract(contract Contract) string {
	if contract.Schema != ContractSchema || contract.ID == "" || contract.SourcePath == "" ||
		contract.SourcePackage == "" || contract.SourceNamespace == "" {
		return "CONTRACT_SCHEMA_INVALID"
	}
	if contract.SourceEntities < 1 || contract.SourceActivities < 1 || contract.FixedDenominator != FixedDenominator ||
		contract.Budget != (IntegerSet{InterpretationCandidates: 2, UnresolvedBranches: 1, EvidencePaths: 2}) {
		return "CONTRACT_DENOMINATOR_INVALID"
	}
	if len(contract.Cases) != ExpectedCaseTotal || len(contract.NotClaimed) != 4 {
		return "CONTRACT_CASE_TOTAL_INVALID"
	}
	seen := map[string]bool{}
	classes := map[string]bool{}
	for _, spec := range contract.Cases {
		if spec.ID == "" || seen[spec.ID] || classes[spec.Class] || spec.Coordinate.Stage == "" || spec.Coordinate.Step == "" || spec.Coordinate.Reason == "" || spec.Claim.CaseID != spec.ID {
			return "CONTRACT_CASE_ID_INVALID"
		}
		seen[spec.ID], classes[spec.Class] = true, true
		if !validCounts(spec.Counts) || spec.Claim.From == "" || spec.Claim.To == "" || spec.Claim.Reason == "" {
			return "CONTRACT_CASE_COUNTS_INVALID"
		}
		if spec.InputState != "KNOWN" && spec.InputState != "UNKNOWN" {
			return "CONTRACT_INPUT_STATE_INVALID"
		}
	}
	for _, class := range []string{"ZERO", "BOUNDARY", "OVER", "UNKNOWN"} {
		if !classes[class] {
			return "CONTRACT_CASE_CLASS_INVALID"
		}
	}
	return ""
}

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || !validSHA(receipt.SubjectSHA) || receipt.ContractID == "" ||
		receipt.Producer != Producer || receipt.Consumer != Consumer || receipt.MetaOperation != MetaOperation ||
		receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || receipt.Reason != "AMBIGUITY_BUDGET_CONTRACT_SATISFIED" {
		return fmt.Errorf("AMBIGUITY_RECEIPT_IDENTITY_INVALID")
	}
	if receipt.Summary.CasesTotal != ExpectedCaseTotal || receipt.Summary.CasesSatisfied != ExpectedCaseTotal ||
		receipt.Summary.CoordinatesTotal != FixedDenominator || receipt.Summary.CoordinatesSatisfied != FixedDenominator ||
		receipt.Summary.FixedDenominator != FixedDenominator || receipt.Summary.ZeroAmbiguityCases != 1 ||
		receipt.Summary.BoundaryCases != 1 || receipt.Summary.OverBudgetCases != 1 || receipt.Summary.UnknownCases != 1 ||
		receipt.Summary.LowerResolutionCases != 2 || receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority ||
		len(receipt.Cases) != ExpectedCaseTotal || len(receipt.Claims) != ExpectedCaseTotal || len(receipt.Indicators) != FixedDenominator ||
		len(receipt.Proofs) != 3 || len(receipt.NotClaimed) != 4 {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUMMARY_INVALID")
	}
	if receipt.Coordinate.Stage != "ambiguity-budget" || receipt.Coordinate.Step != "seal-receipt" {
		return fmt.Errorf("AMBIGUITY_RECEIPT_COORDINATE_INVALID")
	}
	ids := make([]string, 0, len(receipt.Cases))
	for index, result := range receipt.Cases {
		if result.Status != "SATISFIED" || result.ID == "" || result.Claim.CaseID != result.ID || result.Coordinate.Stage == "" || result.Coordinate.Step == "" || result.Coordinate.Reason == "" {
			return fmt.Errorf("AMBIGUITY_RECEIPT_CASE_INVALID")
		}
		if receipt.Claims[index] != result.Claim {
			return fmt.Errorf("AMBIGUITY_RECEIPT_CLAIM_TRANSITION_LOST")
		}
		ids = append(ids, result.ID)
	}
	if len(unique(ids)) != len(ids) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_CASE_DUPLICATE")
	}
	for _, indicator := range receipt.Indicators {
		if indicator.Producer != Producer || indicator.Consumer != Consumer || indicator.MetaOperation != MetaOperation || !indicator.Satisfied || indicator.Observed != indicator.Expected {
			return fmt.Errorf("AMBIGUITY_RECEIPT_INDICATOR_INVALID")
		}
	}
	for _, proof := range receipt.Proofs {
		if proof.Producer != Producer || proof.Consumer != Consumer || !proof.Passed || proof.EvidenceDigest == "" {
			return fmt.Errorf("AMBIGUITY_RECEIPT_PROOF_INVALID")
		}
	}
	if receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_DIGEST_INVALID")
	}
	return nil
}

func unique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	output := result[:0]
	for _, value := range result {
		if len(output) == 0 || output[len(output)-1] != value {
			output = append(output, value)
		}
	}
	return output
}
