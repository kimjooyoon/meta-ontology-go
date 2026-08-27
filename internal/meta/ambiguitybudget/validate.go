package ambiguitybudget

import (
	"fmt"
	"sort"
	"strings"
)

func Validate(receipt Receipt) error {
	if receipt.Schema != ReceiptSchema || !validSHA(receipt.SubjectSHA) || receipt.ContractID == "" ||
		receipt.Producer != Producer || receipt.Consumer != Consumer || receipt.MetaOperation != MetaOperation ||
		receipt.ConformanceDecision != "PASS" || receipt.ConformanceResolution != "EXACT" ||
		receipt.ConformanceReason != "CONFORMANCE_CASES_MATCHED" || receipt.SubjectDecision != "MIXED" ||
		receipt.SubjectResolution != "LOWER_RESOLUTION" || receipt.SubjectReason == "" {
		return fmt.Errorf("AMBIGUITY_RECEIPT_IDENTITY_INVALID")
	}
	if !validPolicy(receipt.BudgetPolicy) || receipt.BudgetBinding != budgetBinding(receipt.BudgetPolicy) ||
		receipt.BudgetAuthority != "CONTRACT_POLICY" || !validEffects(receipt.Effects) ||
		len(receipt.Cases) != ExpectedCaseTotal || len(receipt.Claims) != ExpectedCaseTotal ||
		len(receipt.Indicators) != ExpectedCaseTotal*IntegerDimensions || len(receipt.Interventions) != ExpectedInterventions ||
		len(receipt.Proofs) != 3 || len(receipt.NotClaimed) != 4 {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUMMARY_INVALID")
	}
	if receipt.Summary.Denominator != expectedDenominator() || receipt.Summary.IntegerDimensions != IntegerDimensions ||
		receipt.Summary.CasesTotal != ExpectedCaseTotal || receipt.Summary.Numerator != summarize(receipt.Cases, receipt.Interventions, receipt.Summary.Denominator).Numerator {
		return fmt.Errorf("AMBIGUITY_RECEIPT_DENOMINATOR_INVALID")
	}
	if receipt.SubjectCoordinate.Stage != "ambiguity-budget" || receipt.SubjectCoordinate.Step != "subject-resolution" || receipt.SubjectCoordinate.Reason != receipt.SubjectReason {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUBJECT_COORDINATE_INVALID")
	}
	if receipt.Source.Lowering != canonicalLowering || !validDigest(receipt.Source.Digest) || !validDigest(receipt.Source.SemanticDigest) ||
		len(receipt.Source.Programs) != ExpectedCaseTotal+1 || receipt.Source.Activities != ExpectedSourceActivities || receipt.Source.Entities != ExpectedSourceEntities {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SOURCE_INVALID")
	}

	ids := make([]string, 0, len(receipt.Cases))
	for index, result := range receipt.Cases {
		if err := validateCase(result, receipt.BudgetPolicy); err != nil {
			return err
		}
		if receipt.Claims[index] != result.Claim {
			return fmt.Errorf("AMBIGUITY_RECEIPT_CLAIM_LEDGER_INVALID")
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
			intervention.ClaimBefore.Proposition == "" || intervention.ClaimAfter.Proposition == "" ||
			!validDigest(intervention.ClaimBefore.PropositionDigest) || !validDigest(intervention.ClaimAfter.PropositionDigest) ||
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

func validateCase(result CaseReceipt, policy BudgetPolicy) error {
	if result.ID == "" || result.Activity == "" || result.Program == "" || !validDigest(result.RawSourceDigest) ||
		!validDigest(result.ProgramDigest) || !validDigest(result.ProgramSemanticDigest) || !validDigest(result.ActivitySemanticDigest) ||
		!validDigest(result.ElementDigest) || !validDigest(result.PropositionDigest) || !validDigest(result.EvidenceDigest) ||
		result.ElementDigest != digestValue(result.Elements) || result.Counts != elementCounts(result.Elements) ||
		result.Proposition != proposition(result.ID, policy) || result.PropositionDigest != digestBytes([]byte(result.Proposition)) ||
		result.Claim.CaseID != result.ID || result.Claim.Proposition != result.Proposition || result.Claim.PropositionDigest != result.PropositionDigest ||
		result.Claim.From != "OPEN" || result.Claim.Stage != result.Coordinate.Stage || result.Claim.Step != result.Coordinate.Step ||
		result.Claim.Reason != result.Reason || result.Claim.EvidenceDigest != result.EvidenceDigest || result.Claim.EvidenceDigest == "" {
		return fmt.Errorf("AMBIGUITY_RECEIPT_CASE_INVALID")
	}
	program := ProgramObservation{ID: result.ID, Activity: result.Activity, Program: result.Program, Counts: result.Counts,
		UnobservedDimensions: result.UnobservedDimensions}
	if result.Class != derivedClass(program, policyCounts(policy)) || result.InputState != inputState(program) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_DERIVED_CLASS_INVALID")
	}
	wantDecision, wantResolution, wantReason := subjectDecision(program, policyCounts(policy))
	if result.Decision != wantDecision || result.Resolution != wantResolution || result.Reason != wantReason || result.Claim.To != claimTarget(result.Decision) {
		return fmt.Errorf("AMBIGUITY_RECEIPT_SUBJECT_DECISION_INVALID")
	}
	if len(result.UnobservedDimensions) > 0 {
		if len(result.UnobservedDimensions) != 1 || result.UnobservedDimensions[0] != "unresolved_branches" ||
			result.Coordinate != (Coordinate{Stage: "AMBIGUITY_OBSERVATION", Step: "unresolved_branches", Reason: "AMBIGUITY_COORDINATE_UNOBSERVED"}) ||
			len(result.ObservationGaps) != 1 || result.ObservationGaps[0].Coordinate != result.Coordinate {
			return fmt.Errorf("AMBIGUITY_RECEIPT_UNKNOWN_COORDINATE_INVALID")
		}
	} else if result.Coordinate.Stage != "AMBIGUITY_BUDGET" || result.Coordinate.Step != "case:"+result.ID || result.Coordinate.Reason != result.Reason {
		return fmt.Errorf("AMBIGUITY_RECEIPT_CLAIM_COORDINATE_INVALID")
	}
	return nil
}

func elementCounts(elements AmbiguityElements) IntegerSet {
	return IntegerSet{InterpretationCandidates: len(elements.CandidateIDs), UnresolvedBranches: len(elements.UnresolvedBranchIDs), EvidencePaths: len(elements.EvidencePathIDs)}
}

func expectedDenominator() Denominator {
	return Denominator{Schema: DenominatorSchema, Version: "v1", Cases: ExpectedCaseTotal,
		IntegerObservations: ExpectedCaseTotal * IntegerDimensions, Claims: ExpectedCaseTotal,
		Interventions: ExpectedInterventions, AuthorityObservations: 1}
}

func validEffects(value Effects) bool {
	return value.Schema == EffectsSchema && value.Version == "v1" && validDigest(value.ArtifactDigest) && value.TrackedAndUntracked &&
		validDigest(value.SnapshotBeforeDigest) && validDigest(value.SnapshotAfterDigest) && value.RepositoryWrites == 0 &&
		value.WriteSetEqual && value.MutationAuthority == "UNKNOWN" && value.MutationAuthorityResolution == "NOT_OBSERVED"
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
