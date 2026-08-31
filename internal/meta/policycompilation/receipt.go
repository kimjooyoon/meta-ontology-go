package policycompilation

import (
	"errors"
	"fmt"
	"sort"
)

type claimAssessment struct {
	Predicate string
	Outcome   string
	Observed  bool
	Stage     string
	Step      int
	Reason    string
}

func BuildReceipt(policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case, generated, independent []DecisionResult, writeSet WriteSetObservation, publicCLI PublicCLIEvidence) (Receipt, error) {
	if len(cases) != ExpectedCaseCount || len(cases) != len(generated) || len(cases) != len(independent) {
		return Receipt{}, fmt.Errorf("canonical case denominator must be %d and execution counts must agree", ExpectedCaseCount)
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return Receipt{}, err
	}
	if err := validateClaimPredicates(policy.Rules); err != nil {
		return Receipt{}, err
	}
	metaOperation, proofChoice := sourceReceiptBindings(policy)
	order := make([]int, len(cases))
	for index := range cases {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return cases[order[i]].ID < cases[order[j]].ID })
	receipt := Receipt{
		Schema: ReceiptSchema, Policy: policy, MetaOperation: metaOperation,
		ProofChoice: proofChoice, GeneratedDigest: judgeHash,
		Cases:        make([]CaseReceipt, 0, len(cases)),
		Summary:      CaseSummary{CaseCount: len(cases), CanonicalCaseDenominator: ExpectedCaseCount},
		Claims:       ClaimLedger{Schema: ClaimLedgerSchema, Events: make([]ClaimTransition, 0, len(cases)*ClaimPredicateCount*2)},
		Verification: Verification{Decision: VerificationPass, ConformanceDecision: VerificationPass, SubjectResolution: SubjectUnresolved, IndependentReplayed: true, GeneratedReplayed: true, LedgerVerified: true, FixedDenominator: policy.Denominator, CaseDenominator: len(cases)},
		Evidence:     make([]EvidenceObservation, 0, len(cases)), WriteSet: writeSet,
		CurrentEvidence: CurrentEvidence{Class: EvidenceCurrent, Provenance: "producer runner-temp artifact observation", SourceDigest: policy.SourceDigest, ArtifactSourceDigest: artifact.Policy.SourceDigest, GeneratedJudgeDigest: judgeHash, IndependentDigest: policy.SemanticDigest},
		PublicCLI:       publicCLI,
	}
	prior := ""
	for _, index := range order {
		input, generatedResult, independentResult := cases[index], generated[index], independent[index]
		sourceResult := EvaluateSourcePolicy(policy, input)
		observation := ObservationDigest(input)
		stored := CaseReceipt{ID: input.ID, ValidatorExpectation: input.ValidatorExpectation, EvidenceClass: input.EvidenceClass, ObservationDigest: observation, Provenance: input.Provenance, Source: sourceResult, Generated: generatedResult, Independent: independentResult, ClaimStartDigest: prior}
		receipt.Evidence = append(receipt.Evidence, evidenceForCase(input, policy))
		assessments := assessClaimPredicates(policy, artifact, judgeHash, input, sourceResult, generatedResult, independentResult)
		stored.ClaimPredicates = make([]ClaimPredicateObservation, 0, ClaimPredicateCount)
		for _, rule := range policy.Rules {
			assessment := assessments[rule.Claim]
			id := claimID(input.ID, rule)
			stored.ClaimPredicates = append(stored.ClaimPredicates, ClaimPredicateObservation{ClaimID: id, Predicate: assessment.Predicate, Outcome: assessment.Outcome, Observed: assessment.Observed, Stage: assessment.Stage, Step: assessment.Step, Reason: assessment.Reason, ObservationDigest: observation, Provenance: input.Provenance})
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{ClaimID: id, Predicate: assessment.Predicate, From: ClaimUnrecorded, To: ClaimOpen, Decision: generatedResult.Decision, Stage: rule.Stage, Step: rule.Step, Reason: "CLAIM_OPENED", ObservationDigest: observation, Provenance: input.Provenance}, prior)
		}
		for _, rule := range policy.Rules {
			assessment := assessments[rule.Claim]
			id := claimID(input.ID, rule)
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{ClaimID: id, Predicate: assessment.Predicate, From: ClaimOpen, To: assessment.Outcome, Decision: generatedResult.Decision, Stage: assessment.Stage, Step: assessment.Step, Reason: assessment.Reason, ObservationDigest: observation, Provenance: input.Provenance, Observed: assessment.Observed}, prior)
			switch assessment.Outcome {
			case ClaimDischarged:
				receipt.Summary.ClaimPredicatesDischarged++
			case ClaimRefuted:
				receipt.Summary.ClaimPredicatesRefuted++
			case ClaimOpen:
				receipt.Summary.ClaimPredicatesOpen++
			}
		}
		stored.ClaimEndDigest = prior
		stored.AllDecisionsEquivalent = sameResult(sourceResult, generatedResult) && sameResult(sourceResult, independentResult)
		stored.DecisionsEquivalent = sameResult(generatedResult, independentResult)
		stored.ValidatorExpectationConfirmed = sourceResult.Decision == input.ValidatorExpectation && generatedResult.Decision == input.ValidatorExpectation && independentResult.Decision == input.ValidatorExpectation
		if !stored.AllDecisionsEquivalent || !stored.DecisionsEquivalent || !stored.ValidatorExpectationConfirmed {
			receipt.Verification.Decision, receipt.Verification.ConformanceDecision = VerificationFail, VerificationFail
		}
		switch generatedResult.Decision {
		case DecisionPass:
			receipt.Summary.PassCount++
		case DecisionFailClosed:
			receipt.Summary.FailClosedCount++
		case DecisionUnknown:
			receipt.Summary.UnknownCount++
		default:
			return Receipt{}, fmt.Errorf("case %q has unsupported decision %q", input.ID, generatedResult.Decision)
		}
		if stored.DecisionsEquivalent {
			receipt.Summary.GeneratedIndependentEquivalent++
		}
		if stored.AllDecisionsEquivalent {
			receipt.Summary.SourceAllEquivalent++
		}
		if stored.ValidatorExpectationConfirmed {
			receipt.Summary.ValidatorExpectationsConfirmed++
		}
		receipt.Cases = append(receipt.Cases, stored)
	}
	receipt.Claims.EventCount, receipt.Claims.HeadDigest = len(receipt.Claims.Events), prior
	if err := verifyClaimLedger(receipt.Claims); err != nil {
		return Receipt{}, err
	}
	if receipt.Verification.Decision != VerificationPass || receipt.Summary.GeneratedIndependentEquivalent != ExpectedCaseCount || receipt.Summary.SourceAllEquivalent != ExpectedCaseCount || receipt.Summary.ValidatorExpectationsConfirmed != ExpectedCaseCount {
		return Receipt{}, errors.New("canonical case conformance is incomplete")
	}
	if err := verifyClaimPredicates(receipt, policy); err != nil {
		return Receipt{}, err
	}
	digest, err := ReceiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.ReceiptDigest = digest
	return receipt, nil
}

func FinalizeReceipt(receipt *Receipt) error {
	digest, err := ReceiptDigest(*receipt)
	if err != nil {
		return err
	}
	receipt.ReceiptDigest = digest
	return nil
}

func evidenceForCase(input Case, policy CompiledPolicy) EvidenceObservation {
	return EvidenceObservation{Class: input.EvidenceClass, CaseID: input.ID, SourceDigest: input.ObservedSourceDigest, ArtifactSourceDigest: input.ObservedArtifactSourceDigest, GeneratedJudgeDigest: input.ObservedGeneratedJudgeDigest, IndependentDigest: input.ObservedIndependentDigest, SemanticDigest: policy.SemanticDigest, ObservationDigest: ObservationDigest(input), Provenance: input.Provenance}
}

func claimID(caseID string, rule Rule) string {
	return fmt.Sprintf("gooo://meta-policy-compilation/claim/%s/%02d-%s", caseID, rule.Step, rule.Claim)
}

func sourceReceiptBindings(policy CompiledPolicy) (string, string) {
	metaOperation, proofChoice := "", ""
	for _, rule := range policy.Rules {
		if rule.Claim == ClaimPredicateDecisionReduction {
			metaOperation = rule.MetaOperation
		}
		if rule.Claim == ClaimPredicateProofSelection {
			proofChoice = rule.ProofChoice
		}
	}
	return metaOperation, proofChoice
}

func assessClaimPredicates(policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, input Case, source, generated, independent DecisionResult) map[string]claimAssessment {
	assessments := make(map[string]claimAssessment, ClaimPredicateCount)
	for _, rule := range policy.Rules {
		assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: rule.Stage, Step: rule.Step, Reason: ConditionEvidenceUnavailable}
		switch rule.Claim {
		case ClaimPredicateSourceBound:
			assessment = assessDigest(rule, input.ObservedSourceDigest, policy.SourceDigest, "SOURCE_DIGEST_MISMATCH")
		case ClaimPredicateArtifactBound:
			assessment = assessDigest(rule, input.ObservedArtifactSourceDigest, policy.SourceDigest, "ARTIFACT_SOURCE_MISMATCH")
		case ClaimPredicateGeneratedExecution:
			assessment = assessGenerated(rule, input, generated, judgeHash, policy)
		case ClaimPredicateIndependentReplay:
			assessment = assessIndependent(rule, input, source, independent, policy)
		case ClaimPredicateProofSelection:
			if rule.ProofChoice != "" {
				assessment = discharged(rule, "PROOF_SELECTION_OBSERVED")
			}
		case ClaimPredicateLedgerChain:
			assessment = discharged(rule, "LEDGER_CHAIN_APPENDED")
		case ClaimPredicateDecisionReduction:
			assessment = assessReduction(rule, policy, input, source, generated, independent)
		case ClaimPredicateLineageSeal:
			assessment = assessLineage(rule, policy, artifact, judgeHash, input, source, generated, independent)
		}
		assessments[rule.Claim] = assessment
	}
	return assessments
}

func assessDigest(rule Rule, observed, expected, mismatch string) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionEvidenceUnavailable}
	if observed == "" {
		assessment.Reason = ConditionDigestUnavailable
		return assessment
	}
	if !ValidDigest(observed) {
		assessment.Reason = ConditionMalformedDigest
		return assessment
	}
	assessment.Stage = rule.Stage
	if observed != expected {
		assessment.Outcome, assessment.Observed, assessment.Reason = ClaimRefuted, true, mismatch
		return assessment
	}
	return discharged(rule, "DIGEST_BOUND")
}

func assessGenerated(rule Rule, input Case, value DecisionResult, judgeHash string, policy CompiledPolicy) claimAssessment {
	if input.ObservedGeneratedJudgeDigest == "" {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionDigestUnavailable}
	}
	if !ValidDigest(input.ObservedGeneratedJudgeDigest) {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionMalformedDigest}
	}
	if input.ObservedGeneratedJudgeDigest != judgeHash || !validResult(value, input.ID, policy) {
		return refuted(rule, "GENERATED_EXECUTION_MISMATCH")
	}
	return discharged(rule, "GENERATED_EXECUTION_OBSERVED")
}

func assessIndependent(rule Rule, input Case, source, independent DecisionResult, policy CompiledPolicy) claimAssessment {
	if input.ObservedIndependentDigest == "" {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionDigestUnavailable}
	}
	if !ValidDigest(input.ObservedIndependentDigest) {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionMalformedDigest}
	}
	if !validResult(independent, input.ID, policy) || !sameResult(source, independent) || input.ObservedIndependentDigest != policy.SemanticDigest {
		return refuted(rule, "INDEPENDENT_REPLAY_MISMATCH")
	}
	return discharged(rule, "INDEPENDENT_REPLAY_OBSERVED")
}

func assessReduction(rule Rule, policy CompiledPolicy, input Case, source, generated, independent DecisionResult) claimAssessment {
	if !validResult(source, input.ID, policy) || !sameResult(source, generated) || !sameResult(source, independent) || !reductionResultMatches(policy, input, source) {
		return refuted(rule, "DECISION_REDUCTION_MISMATCH")
	}
	return discharged(rule, source.Reason)
}

func assessLineage(rule Rule, policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, input Case, source, generated, independent DecisionResult) claimAssessment {
	if input.ObservedSourceDigest == "" || input.ObservedArtifactSourceDigest == "" || input.ObservedGeneratedJudgeDigest == "" || input.ObservedIndependentDigest == "" {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionDigestUnavailable}
	}
	if !ValidDigest(input.ObservedSourceDigest) || !ValidDigest(input.ObservedArtifactSourceDigest) || !ValidDigest(input.ObservedGeneratedJudgeDigest) || !ValidDigest(input.ObservedIndependentDigest) {
		return claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: rule.Step, Reason: ConditionMalformedDigest}
	}
	if input.ObservedSourceDigest != policy.SourceDigest || input.ObservedArtifactSourceDigest != artifact.Policy.SourceDigest || input.ObservedGeneratedJudgeDigest != judgeHash || input.ObservedIndependentDigest != policy.SemanticDigest {
		return refuted(rule, "LINEAGE_DIGEST_MISMATCH")
	}
	if !sameResult(source, generated) || !sameResult(source, independent) {
		return refuted(rule, "LINEAGE_RESULT_MISMATCH")
	}
	return discharged(rule, "LINEAGE_SEALED")
}

func discharged(rule Rule, reason string) claimAssessment {
	return claimAssessment{Predicate: rule.Claim, Outcome: ClaimDischarged, Observed: true, Stage: rule.Stage, Step: rule.Step, Reason: reason}
}

func refuted(rule Rule, reason string) claimAssessment {
	return claimAssessment{Predicate: rule.Claim, Outcome: ClaimRefuted, Observed: true, Stage: rule.Stage, Step: rule.Step, Reason: reason}
}

func reductionResultMatches(policy CompiledPolicy, input Case, result DecisionResult) bool {
	for _, row := range policy.Reduction.Rules {
		if sourceConditionMatches(row.Condition, policy.SourceDigest, policy.SemanticDigest, input) {
			return result.Decision == row.Decision && result.Stage == row.Stage && result.Step == row.Step && result.Reason == row.Reason && result.UnknownClass == row.UnknownClass && result.NextOperation == row.NextOperation && sameStrings(result.BlockedBy, row.BlockedBy)
		}
	}
	return false
}

func validResult(value DecisionResult, caseID string, policy CompiledPolicy) bool {
	if value.CaseID != caseID || value.PolicyDigest != policy.SourceDigest || value.SemanticDigest != policy.SemanticDigest || value.Denominator != policy.Denominator || value.Stage == "" || value.Reason == "" {
		return false
	}
	if value.Decision != DecisionPass && value.Decision != DecisionFailClosed && value.Decision != DecisionUnknown {
		return false
	}
	if value.Decision == DecisionUnknown {
		return value.Step > 0 && value.UnknownClass != "" && value.NextOperation != "" && len(value.BlockedBy) > 0
	}
	return value.UnknownClass == "" && value.NextOperation == "" && len(value.BlockedBy) == 0
}

func sameResult(left, right DecisionResult) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision && left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation && sameStrings(left.BlockedBy, right.BlockedBy) && left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest && left.Denominator == right.Denominator
}

func sameStrings(left, right []string) bool {
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

func appendClaimEvent(ledger *ClaimLedger, event ClaimTransition, prior string) string {
	event.Event, event.PriorDigest = len(ledger.Events)+1, prior
	canonical := event
	canonical.Digest = ""
	digest, _ := digestJSON(canonical)
	event.Digest = digest
	ledger.Events = append(ledger.Events, event)
	return digest
}
