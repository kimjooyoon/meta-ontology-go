package ambiguitybudget

import (
	"fmt"
	"sort"
	"strings"
)

func validateContract(contract Contract) string {
	if contract.Schema != ContractSchema || contract.ID == "" || contract.SourcePath == "" ||
		contract.SourcePackage == "" || contract.SourceNamespace == "" || contract.BudgetActivity == "" {
		return "CONTRACT_SCHEMA_INVALID"
	}
	if contract.FixedDenominator != FixedDenominator {
		return "CONTRACT_DENOMINATOR_INVALID"
	}
	if len(contract.Cases) != ExpectedCaseTotal || len(contract.Interventions) != ExpectedInterventions || len(contract.NotClaimed) != 4 {
		return "CONTRACT_CARDINALITY_INVALID"
	}
	caseIDs, activities := map[string]bool{}, map[string]bool{}
	for _, item := range contract.Cases {
		if item.ID == "" || item.Activity == "" || caseIDs[item.ID] || activities[item.Activity] {
			return "CONTRACT_CASE_ID_INVALID"
		}
		caseIDs[item.ID], activities[item.Activity] = true, true
	}
	interventionIDs := map[string]bool{}
	for _, item := range contract.Interventions {
		if item.ID == "" || item.TargetActivity == "" || interventionIDs[item.ID] ||
			(item.Kind != "SEMANTIC" && item.Kind != "NONSEMANTIC") {
			return "CONTRACT_INTERVENTION_INVALID"
		}
		interventionIDs[item.ID] = true
	}
	return ""
}

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || !validSHA(receipt.SubjectSHA) || receipt.ContractID == "" ||
		receipt.Producer != Producer || receipt.Consumer != Consumer || receipt.MetaOperation != MetaOperation ||
		receipt.ConformanceDecision != "PASS" || receipt.ConformanceResolution != "EXACT" ||
		receipt.ConformanceReason != "CONFORMANCE_CASES_MATCHED" || receipt.SubjectDecision != "MIXED" ||
		receipt.SubjectResolution != "LOWER_RESOLUTION" || receipt.SubjectReason == "" {
		return fmt.Errorf("AMBIGUITY_RECEIPT_IDENTITY_INVALID")
	}
	if receipt.Budget != expectedBudget() || receipt.Summary != expectedSummary() ||
		receipt.Effects != (Effects{}) || len(receipt.Cases) != ExpectedCaseTotal ||
		len(receipt.Claims) != ExpectedCaseTotal || len(receipt.Indicators) != ExpectedCaseTotal*IntegerDimensions ||
		len(receipt.Interventions) != ExpectedInterventions || len(receipt.Proofs) != 3 || len(receipt.NotClaimed) != 4 {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUMMARY_INVALID")
	}
	if receipt.SubjectCoordinate.Stage != "ambiguity-budget" || receipt.SubjectCoordinate.Step != "subject-resolution" || receipt.SubjectCoordinate.Reason != receipt.SubjectReason {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUBJECT_COORDINATE_INVALID")
	}
	if receipt.Source.Lowering != canonicalLowering || !validDigest(receipt.Source.Digest) || !validDigest(receipt.Source.SemanticDigest) ||
		len(receipt.Source.Programs) != ExpectedCaseTotal+1 || receipt.Source.Activities != ExpectedCaseTotal+1 {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SOURCE_INVALID")
	}
	ids := make([]string, 0, len(receipt.Cases))
	for index, result := range receipt.Cases {
		if result.ID == "" || result.Activity == "" || result.Class == "" || result.Program == "" ||
			!validCounts(result.Counts, result.UnobservedDimensions) || !validDigest(result.ProgramDigest) || !validDigest(result.EvidenceDigest) ||
			result.Conformance != "MATCH" || result.Coordinate.Stage == "" || result.Coordinate.Step == "" ||
			result.Claim.CaseID != result.ID || result.Claim.From != "OPEN" ||
			result.Claim.Stage != result.Coordinate.Stage || result.Claim.Step != result.Coordinate.Step ||
			result.Claim.Reason != result.Reason || result.Claim.EvidenceDigest != result.EvidenceDigest ||
			result.Claim.EvidenceDigest == "" || receipt.Claims[index] != result.Claim {
			return fmt.Errorf("AMBIGUITY_RECEIPT_CASE_INVALID")
		}
		if len(result.UnobservedDimensions) > 0 {
			if result.Class != "UNKNOWN" || result.InputState != "UNKNOWN" || result.Decision != "UNKNOWN" ||
				result.Resolution != "LOWER_RESOLUTION" || result.Reason != "AMBIGUITY_COORDINATE_UNOBSERVED" ||
				result.Coordinate.Stage != "AMBIGUITY_OBSERVATION" || len(result.UnobservedDimensions) != 1 ||
				result.Coordinate.Step != result.UnobservedDimensions[0] || result.Coordinate.Reason != result.Reason || result.Claim.To != "OPEN" {
				return fmt.Errorf("AMBIGUITY_RECEIPT_UNKNOWN_TRANSITION_INVALID")
			}
		} else if result.InputState != "KNOWN" || result.Class != derivedClass(computesProgram{Counts: result.Counts}, receipt.Budget) ||
			result.Coordinate.Stage != "AMBIGUITY_BUDGET" || result.Coordinate.Step != "case:"+result.ID ||
			result.Coordinate.Reason != result.Reason || result.Claim.To != claimTarget(result.Decision) {
			return fmt.Errorf("AMBIGUITY_RECEIPT_CLAIM_TRANSITION_INVALID")
		}
		ids = append(ids, result.ID)
	}
	if len(unique(ids)) != len(ids) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_CASE_DUPLICATE")
	}
	for _, indicator := range receipt.Indicators {
		if indicator.CaseID == "" || indicator.Dimension == "" || indicator.ProofChoice == "" ||
			indicator.Producer != Producer || indicator.Consumer != Consumer || indicator.MetaOperation != MetaOperation ||
			indicator.Relation != "<=" || indicator.Evaluation == "" || !validDigest(indicator.EvidenceDigest) ||
			(!indicator.CoordinateObserved && indicator.Evaluation != "UNOBSERVED") ||
			(indicator.CoordinateObserved && indicator.Evaluation == "UNOBSERVED") {
			return fmt.Errorf("AMBIGUITY_RECEIPT_INDICATOR_INVALID")
		}
	}
	for _, intervention := range receipt.Interventions {
		if intervention.ID == "" || intervention.Kind == "" || intervention.TargetActivity == "" || !intervention.Satisfied ||
			!validDigest(intervention.SourceDigestBefore) || !validDigest(intervention.SourceDigestAfter) ||
			!validDigest(intervention.SemanticDigestBefore) || !validDigest(intervention.SemanticDigestAfter) ||
			!validDigest(intervention.EvidenceDigest) || intervention.ClaimBefore.From != "OPEN" || intervention.ClaimAfter.From != "OPEN" ||
			!validDigest(intervention.ClaimBefore.EvidenceDigest) || !validDigest(intervention.ClaimAfter.EvidenceDigest) {
			return fmt.Errorf("AMBIGUITY_RECEIPT_INTERVENTION_INVALID")
		}
	}
	for _, proof := range receipt.Proofs {
		if proof.Producer != Producer || proof.Consumer != Consumer || proof.MetaOperation == "" ||
			proof.Choice == "" || !proof.Passed || !validDigest(proof.EvidenceDigest) {
			return fmt.Errorf("AMBIGUITY_RECEIPT_PROOF_INVALID")
		}
	}
	if !validDigest(receipt.FactsDigest) || !validDigest(receipt.Digest) || receipt.Digest != receiptDigest(receipt) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_DIGEST_INVALID")
	}
	return nil
}

func expectedSummary() Summary {
	return Summary{CasesTotal: ExpectedCaseTotal, KnownCases: 3, ZeroAmbiguityCases: 1, BoundaryCases: 1,
		OverBudgetCases: 1, UnknownCases: 1, LowerResolutionCases: 2, OpenClaims: 1,
		IntegerDimensions: IntegerDimensions, InterventionsTotal: ExpectedInterventions, FixedDenominator: FixedDenominator}
}

func validSHA(value string) bool {
	return len(value) == 40 && strings.Trim(value, "0123456789abcdefABCDEF") == ""
}

func validDigest(value string) bool {
	return len(value) == len("sha256:")+64 && strings.HasPrefix(value, "sha256:") && strings.Trim(value[len("sha256:"):], "0123456789abcdef") == ""
}

func unique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	out := result[:0]
	for _, value := range result {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}
