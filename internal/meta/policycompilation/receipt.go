package policycompilation

import (
	"encoding/json"
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

func BuildReceipt(policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case, generated, independent []DecisionResult, writeSet WriteSetObservation) (Receipt, error) {
	if len(cases) != len(generated) || len(cases) != len(independent) {
		return Receipt{}, errors.New("case and execution result counts differ")
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return Receipt{}, err
	}
	if err := validateClaimPredicates(policy.Rules); err != nil {
		return Receipt{}, err
	}
	producer := ProducerEvidence{
		Role: "PRODUCER", Stage: "PRODUCE", Step: 1, Reason: "SOURCE_BOUND",
		SourceDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest,
		Denominator: policy.Denominator,
	}
	consumer := ConsumerEvidence{
		Role: "CONSUMER", Stage: "CONSUME", Step: 2, Reason: "ARTIFACT_BOUND",
		ArtifactSourceDigest: artifact.Policy.SourceDigest,
		ArtifactDigest:       artifactDigest(artifact), SourceMatches: artifact.Policy.SourceDigest == policy.SourceDigest,
		RulesMatch: len(artifact.Policy.Rules) == policy.Denominator,
	}
	metaOperation, proofChoice := sourceReceiptBindings(policy)
	order := make([]int, len(cases))
	for index := range cases {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return cases[order[i]].ID < cases[order[j]].ID })
	receipt := Receipt{
		Schema: ReceiptSchema, Policy: policy, Producer: producer, Consumer: consumer,
		MetaOperation: metaOperation, ProofChoice: proofChoice, GeneratedDigest: judgeHash,
		Cases:        make([]CaseReceipt, 0, len(cases)),
		Summary:      CaseSummary{CaseCount: len(cases)},
		Claims:       ClaimLedger{Schema: ClaimLedgerSchema, Events: make([]ClaimTransition, 0, len(cases)*ClaimPredicateCount*2)},
		Verification: Verification{Decision: VerificationPass, ConformanceDecision: VerificationPass, SubjectResolution: SubjectUnresolved, IndependentReplayed: true, GeneratedReplayed: true, LedgerVerified: true, FixedDenominator: policy.Denominator, CaseDenominator: len(cases)},
		Evidence:     make([]EvidenceObservation, 0, len(cases)), WriteSet: writeSet,
	}
	prior := ""
	for _, index := range order {
		input, generatedResult, independentResult := cases[index], generated[index], independent[index]
		sourceResult := EvaluateSourcePolicy(policy, input)
		observationDigest := observationDigest(input)
		caseReceipt := CaseReceipt{
			ID: input.ID, ValidatorExpectation: input.ValidatorExpectation, EvidenceClass: input.EvidenceClass,
			ObservationDigest: observationDigest, Provenance: input.Provenance,
			Source: sourceResult, Generated: generatedResult, Independent: independentResult,
		}
		receipt.Evidence = append(receipt.Evidence, evidenceForCase(input, artifact, policy))
		caseReceipt.ClaimStartDigest = prior
		assessments := assessClaimPredicates(policy, artifact, judgeHash, input, sourceResult, generatedResult, independentResult)
		caseReceipt.ClaimPredicates = make([]ClaimPredicateObservation, 0, len(assessments))
		for _, rule := range policy.Rules {
			assessment := assessments[rule.Claim]
			claimID := claimID(input.ID, rule)
			caseReceipt.ClaimPredicates = append(caseReceipt.ClaimPredicates, ClaimPredicateObservation{
				ClaimID: claimID, Predicate: assessment.Predicate, Outcome: assessment.Outcome,
				Observed: assessment.Observed, Stage: assessment.Stage, Step: assessment.Step,
				Reason: assessment.Reason, ObservationDigest: observationDigest, Provenance: input.Provenance,
			})
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{
				ClaimID: claimID, Predicate: assessment.Predicate, From: ClaimUnrecorded, To: ClaimOpen,
				Decision: generatedResult.Decision, Stage: rule.Stage, Step: rule.Step, Reason: "CLAIM_OPENED",
				ObservationDigest: observationDigest, Provenance: input.Provenance, Observed: false,
				PriorDigest: prior,
			}, prior)
		}
		for _, rule := range policy.Rules {
			assessment := assessments[rule.Claim]
			claimID := claimID(input.ID, rule)
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{
				ClaimID: claimID, Predicate: assessment.Predicate, From: ClaimOpen, To: assessment.Outcome,
				Decision: generatedResult.Decision, Stage: assessment.Stage, Step: assessment.Step, Reason: assessment.Reason,
				ObservationDigest: observationDigest, Provenance: input.Provenance, Observed: assessment.Observed,
				PriorDigest: prior,
			}, prior)
			switch assessment.Outcome {
			case ClaimDischarged:
				receipt.Summary.ClaimPredicatesDischarged++
			case ClaimRefuted:
				receipt.Summary.ClaimPredicatesRefuted++
			case ClaimOpen:
				receipt.Summary.ClaimPredicatesOpen++
			}
		}
		caseReceipt.ClaimEndDigest = prior
		caseReceipt.AllDecisionsEquivalent = sameDecision(sourceResult, generatedResult) && sameDecision(sourceResult, independentResult)
		caseReceipt.DecisionsEquivalent = sameDecision(generatedResult, independentResult)
		caseReceipt.ValidatorExpectationConfirmed = sourceResult.Decision == input.ValidatorExpectation && generatedResult.Decision == input.ValidatorExpectation && independentResult.Decision == input.ValidatorExpectation
		if !caseReceipt.AllDecisionsEquivalent || !caseReceipt.DecisionsEquivalent || !caseReceipt.ValidatorExpectationConfirmed {
			receipt.Verification.Decision = VerificationFail
			receipt.Verification.ConformanceDecision = VerificationFail
		}
		switch generatedResult.Decision {
		case DecisionPass:
			receipt.Summary.PassCount++
		case DecisionFailClosed:
			receipt.Summary.FailClosedCount++
		case DecisionUnknown:
			receipt.Summary.UnknownCount++
		}
		if caseReceipt.DecisionsEquivalent {
			receipt.Summary.GeneratedIndependentEqual++
		}
		if caseReceipt.ValidatorExpectationConfirmed {
			receipt.Summary.ValidatorExpectationsConfirmed++
		}
		if caseReceipt.AllDecisionsEquivalent {
			receipt.Summary.SourceAllEquivalent++
		}
		receipt.Cases = append(receipt.Cases, caseReceipt)
	}
	receipt.Claims.EventCount = len(receipt.Claims.Events)
	receipt.Claims.HeadDigest = prior
	if receipt.Summary.CaseCount != ExpectedCaseCount || receipt.Summary.GeneratedIndependentEqual != len(cases) || receipt.Summary.SourceAllEquivalent != len(cases) || receipt.Summary.ValidatorExpectationsConfirmed != len(cases) {
		return Receipt{}, fmt.Errorf("case conformance is incomplete: %#v", receipt.Summary)
	}
	if err := verifyClaimPredicates(receipt, policy); err != nil {
		return Receipt{}, err
	}
	var err error
	receipt.ReceiptDigest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func FinalizeReceipt(receipt *Receipt) error {
	digest, err := receiptDigest(*receipt)
	if err != nil {
		return err
	}
	receipt.ReceiptDigest = digest
	return nil
}

func artifactDigest(artifact PolicyArtifact) string {
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return ""
	}
	return DigestBytes(append(data, '\n'))
}

func observationDigest(input Case) string {
	digest, _ := digestJSON(input)
	return digest
}

func claimID(caseID string, rule Rule) string {
	return fmt.Sprintf("gooo://meta-policy-compilation/claim/%s/%02d-%s", caseID, rule.Step, rule.Claim)
}

func sourceReceiptBindings(policy CompiledPolicy) (string, string) {
	metaOperation, proofChoice := MetaOperation, ProofChoice
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

func evidenceForCase(input Case, artifact PolicyArtifact, policy CompiledPolicy) EvidenceObservation {
	return EvidenceObservation{
		Class: input.EvidenceClass, CaseID: input.ID, ProducerAvailable: input.ProducerAvailable,
		ConsumerAvailable: input.ConsumerAvailable, SourceDigest: input.ObservedSourceDigest,
		ArtifactSourceDigest: input.ObservedArtifactSourceDigest, ArtifactDigest: artifactDigest(artifact),
		GeneratedJudgeDigest: input.ObservedGeneratedJudgeDigest, IndependentDigest: input.ObservedIndependentDigest,
		SemanticDigest: policy.SemanticDigest, ObservationDigest: observationDigest(input), Provenance: input.Provenance,
	}
}

func assessClaimPredicates(policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, input Case, source, generated, independent DecisionResult) map[string]claimAssessment {
	assessments := make(map[string]claimAssessment, ClaimPredicateCount)
	for _, rule := range policy.Rules {
		assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: rule.Stage, Step: rule.Step, Reason: "EVIDENCE_UNAVAILABLE"}
		switch rule.Claim {
		case ClaimPredicateSourceBound:
			assessment = assessDigestBinding(rule.Claim, rule.Stage, rule.Step, input.ProducerAvailable, input.ObservedSourceDigest, policy.SourceDigest, ConditionSourceMismatch)
		case ClaimPredicateArtifactBound:
			assessment = assessDigestBinding(rule.Claim, rule.Stage, rule.Step, input.ConsumerAvailable, input.ObservedArtifactSourceDigest, policy.SourceDigest, ConditionArtifactMismatch)
		case ClaimPredicateGeneratedExecution:
			assessment = assessGeneratedExecution(rule, input, generated, judgeHash, policy)
		case ClaimPredicateIndependentReplay:
			assessment = assessIndependentReplay(rule, input, source, independent, policy)
		case ClaimPredicateProofSelection:
			assessment = assessProofSelection(rule, policy)
		case ClaimPredicateLedgerChain:
			assessment = claimAssessment{Predicate: rule.Claim, Outcome: ClaimDischarged, Observed: true, Stage: rule.Stage, Step: rule.Step, Reason: "LEDGER_CHAIN_APPENDED"}
		case ClaimPredicateDecisionReduction:
			assessment = assessDecisionReduction(rule, policy, input, source, generated, independent)
		case ClaimPredicateLineageSeal:
			assessment = assessLineage(rule, policy, artifact, judgeHash, input, source, generated, independent)
		}
		assessments[rule.Claim] = assessment
	}
	return assessments
}

func assessDigestBinding(predicate, stage string, step int, available bool, observed, expected, mismatchReason string) claimAssessment {
	assessment := claimAssessment{Predicate: predicate, Outcome: ClaimOpen, Stage: "VERIFY", Step: 4, Reason: ConditionEvidenceUnavailable}
	if !available {
		return assessment
	}
	if observed == "" {
		assessment.Reason = ConditionDigestUnavailable
		return assessment
	}
	if !ValidDigest(observed) {
		assessment.Reason, assessment.Stage = ConditionMalformedDigest, "LOWER_RESOLUTION"
		return assessment
	}
	assessment.Stage, assessment.Step = stage, step
	if observed == expected {
		assessment.Outcome, assessment.Observed, assessment.Reason = ClaimDischarged, true, "DIGEST_BOUND"
		return assessment
	}
	assessment.Outcome, assessment.Observed, assessment.Reason = ClaimRefuted, true, mismatchReason
	return assessment
}

func assessGeneratedExecution(rule Rule, input Case, value DecisionResult, judgeHash string, policy CompiledPolicy) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "VERIFY", Step: 4, Reason: ConditionEvidenceUnavailable}
	if !input.ProducerAvailable {
		return assessment
	}
	if input.ObservedGeneratedJudgeDigest == "" {
		assessment.Reason = ConditionDigestUnavailable
		return assessment
	}
	if !ValidDigest(input.ObservedGeneratedJudgeDigest) {
		assessment.Reason, assessment.Stage = ConditionMalformedDigest, "LOWER_RESOLUTION"
		return assessment
	}
	if input.ObservedGeneratedJudgeDigest != judgeHash {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "GENERATED_JUDGE_DIGEST_MISMATCH"
		return assessment
	}
	if !validResult(value, input.ID, policy) {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "GENERATED_EXECUTION_INVALID"
		return assessment
	}
	assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimDischarged, true, rule.Stage, rule.Step, "GENERATED_EXECUTION_OBSERVED"
	return assessment
}

func assessIndependentReplay(rule Rule, input Case, source, independent DecisionResult, policy CompiledPolicy) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "VERIFY", Step: 4, Reason: ConditionEvidenceUnavailable}
	if !input.ConsumerAvailable {
		return assessment
	}
	if input.ObservedIndependentDigest == "" {
		assessment.Reason = ConditionDigestUnavailable
		return assessment
	}
	if !ValidDigest(input.ObservedIndependentDigest) {
		assessment.Reason, assessment.Stage = ConditionMalformedDigest, "LOWER_RESOLUTION"
		return assessment
	}
	if !validResult(independent, input.ID, policy) || !sameDecision(source, independent) {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "INDEPENDENT_REPLAY_MISMATCH"
		return assessment
	}
	if input.ObservedIndependentDigest != policy.SemanticDigest {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, ConditionIndependentMismatch
		return assessment
	}
	assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimDischarged, true, rule.Stage, rule.Step, "INDEPENDENT_REPLAY_OBSERVED"
	return assessment
}

func assessProofSelection(rule Rule, policy CompiledPolicy) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: rule.Stage, Step: rule.Step, Reason: "PROOF_SELECTION_MISSING"}
	for _, candidate := range policy.Rules {
		if candidate.Claim == ClaimPredicateProofSelection && candidate.ProofChoice != "" {
			assessment.Outcome, assessment.Observed, assessment.Reason = ClaimDischarged, true, "PROOF_SELECTION_OBSERVED"
			return assessment
		}
	}
	return assessment
}

func assessDecisionReduction(rule Rule, policy CompiledPolicy, input Case, source, generated, independent DecisionResult) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "VERIFY", Step: 4, Reason: ConditionEvidenceUnavailable}
	if !input.ProducerAvailable {
		return assessment
	}
	if !validResult(source, input.ID, policy) || !sameDecision(source, generated) || !sameDecision(source, independent) || !reductionResultMatches(policy, input, source) {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "DECISION_REDUCTION_MISMATCH"
		return assessment
	}
	assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimDischarged, true, rule.Stage, rule.Step, source.Reason
	return assessment
}

func assessLineage(rule Rule, policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, input Case, source, generated, independent DecisionResult) claimAssessment {
	assessment := claimAssessment{Predicate: rule.Claim, Outcome: ClaimOpen, Stage: "LOWER_RESOLUTION", Step: 4, Reason: ConditionEvidenceUnavailable}
	if !input.ProducerAvailable || !input.ConsumerAvailable {
		return assessment
	}
	if hasEmptyDigest(input) {
		assessment.Reason = ConditionDigestUnavailable
		return assessment
	}
	if hasMalformedDigest(input) {
		assessment.Reason = ConditionMalformedDigest
		return assessment
	}
	if input.ObservedSourceDigest != policy.SourceDigest {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, ConditionSourceMismatch
		return assessment
	}
	if input.ObservedArtifactSourceDigest != artifact.Policy.SourceDigest || input.ObservedGeneratedJudgeDigest != judgeHash || input.ObservedIndependentDigest != policy.SemanticDigest {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "LINEAGE_DIGEST_MISMATCH"
		return assessment
	}
	if !validResult(source, input.ID, policy) || !sameDecision(source, generated) || !sameDecision(source, independent) {
		assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimRefuted, true, rule.Stage, rule.Step, "LINEAGE_RESULT_MISMATCH"
		return assessment
	}
	assessment.Outcome, assessment.Observed, assessment.Stage, assessment.Step, assessment.Reason = ClaimDischarged, true, rule.Stage, rule.Step, "LINEAGE_SEALED"
	return assessment
}

func reductionResultMatches(policy CompiledPolicy, input Case, result DecisionResult) bool {
	for _, row := range policy.Reduction.Rules {
		if sourceConditionMatches(row.Condition, policy.SourceDigest, policy.SemanticDigest, input) {
			return result.Decision == row.Decision && result.Stage == row.Stage && result.Step == row.Step && result.Reason == row.Reason
		}
	}
	return false
}

func validResult(value DecisionResult, caseID string, policy CompiledPolicy) bool {
	return value.CaseID == caseID && value.PolicyDigest == policy.SourceDigest && value.SemanticDigest == policy.SemanticDigest && value.Denominator == policy.Denominator && (value.Decision == DecisionPass || value.Decision == DecisionFailClosed || value.Decision == DecisionUnknown) && value.Stage != "" && value.Step > 0 && value.Reason != ""
}

func verifyClaimPredicates(receipt Receipt, policy CompiledPolicy) error {
	if len(receipt.Cases) == 0 {
		return errors.New("claim predicate observations are empty")
	}
	for _, stored := range receipt.Cases {
		if len(stored.ClaimPredicates) != ClaimPredicateCount {
			return fmt.Errorf("case %q does not have %d distinct claim predicates", stored.ID, ClaimPredicateCount)
		}
		seen := make(map[string]bool, ClaimPredicateCount)
		for _, predicate := range stored.ClaimPredicates {
			if seen[predicate.Predicate] || predicate.ClaimID == "" || predicate.ObservationDigest != stored.ObservationDigest || predicate.Provenance != stored.Provenance || (predicate.Outcome != ClaimOpen && predicate.Outcome != ClaimDischarged && predicate.Outcome != ClaimRefuted) {
				return fmt.Errorf("case %q has an invalid claim predicate observation", stored.ID)
			}
			seen[predicate.Predicate] = true
		}
		for _, rule := range policy.Rules {
			if !seen[rule.Claim] {
				return fmt.Errorf("case %q is missing claim predicate %q", stored.ID, rule.Claim)
			}
		}
	}
	return nil
}

func appendClaimEvent(ledger *ClaimLedger, event ClaimTransition, prior string) string {
	event.Event = len(ledger.Events) + 1
	event.PriorDigest = prior
	canonical := event
	canonical.Digest = ""
	digest, _ := digestJSON(canonical)
	event.Digest = digest
	ledger.Events = append(ledger.Events, event)
	return digest
}
