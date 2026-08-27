package policycompilation

import (
	"errors"
	"fmt"
	"sort"
)

func BuildReceipt(policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case, generated, independent []DecisionResult, writeSet WriteSetObservation) (Receipt, error) {
	if len(cases) != len(generated) || len(cases) != len(independent) {
		return Receipt{}, errors.New("case and execution result counts differ")
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
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

	order := make([]int, len(cases))
	for index := range cases {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return cases[order[i]].ID < cases[order[j]].ID })
	receipt := Receipt{
		Schema: ReceiptSchema, Policy: policy, Producer: producer, Consumer: consumer,
		MetaOperation: MetaOperation, ProofChoice: ProofChoice, GeneratedDigest: judgeHash,
		Cases:        make([]CaseReceipt, 0, len(cases)),
		Summary:      CaseSummary{CaseCount: len(cases)},
		Claims:       ClaimLedger{Schema: ClaimLedgerSchema, Events: make([]ClaimTransition, 0, len(cases)*FixedDenominator*2)},
		Verification: Verification{Decision: VerificationPass, ConformanceDecision: VerificationPass, SubjectResolution: SubjectUnresolved, IndependentReplayed: true, GeneratedReplayed: true, LedgerVerified: true, FixedDenominator: policy.Denominator, CaseDenominator: len(cases)},
		Evidence:     make([]EvidenceObservation, 0, len(cases)), WriteSet: writeSet,
	}
	prior := ""
	for _, index := range order {
		input, generatedResult, independentResult := cases[index], generated[index], independent[index]
		sourceResult := EvaluateSourcePolicy(policy, input)
		observationDigest := observationDigest(input)
		caseReceipt := CaseReceipt{ID: input.ID, ValidatorExpectation: input.ValidatorExpectation, EvidenceClass: input.EvidenceClass, ObservationDigest: observationDigest, Provenance: input.Provenance, Source: sourceResult, Generated: generatedResult, Independent: independentResult}
		receipt.Evidence = append(receipt.Evidence, EvidenceObservation{Class: input.EvidenceClass, CaseID: input.ID, ProducerAvailable: input.ProducerAvailable, ConsumerAvailable: input.ConsumerAvailable, SourceDigest: input.ObservedSourceDigest, ArtifactDigest: artifactDigest(artifact), ObservationDigest: observationDigest, Provenance: input.Provenance})
		caseReceipt.ClaimStartDigest = prior
		for _, rule := range policy.Rules {
			claimID := claimID(input.ID, rule)
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{
				ClaimID: claimID, From: ClaimUnrecorded, To: ClaimOpen,
				Decision: generatedResult.Decision, Stage: rule.Stage, Step: rule.Step, Reason: "CLAIM_OPENED", ObservationDigest: observationDigest, Provenance: input.Provenance, PriorDigest: prior,
			}, prior)
		}
		for _, rule := range policy.Rules {
			to, reason := claimOutcome(rule, generatedResult)
			prior = appendClaimEvent(&receipt.Claims, ClaimTransition{
				ClaimID: claimID(input.ID, rule), From: ClaimOpen, To: to,
				Decision: generatedResult.Decision, Stage: rule.Stage, Step: rule.Step, Reason: reason, ObservationDigest: observationDigest, Provenance: input.Provenance, PriorDigest: prior,
			}, prior)
		}
		caseReceipt.ClaimEndDigest = prior
		caseReceipt.AllDecisionsEquivalent = sameDecision(sourceResult, generatedResult) && sameDecision(sourceResult, independentResult)
		caseReceipt.DecisionsEquivalent = sameDecision(generatedResult, independentResult)
		caseReceipt.ValidatorExpectationConfirmed = sourceResult.Decision == input.ValidatorExpectation && generatedResult.Decision == input.ValidatorExpectation && independentResult.Decision == input.ValidatorExpectation
		if !caseReceipt.AllDecisionsEquivalent || !caseReceipt.DecisionsEquivalent || !caseReceipt.ValidatorExpectationConfirmed {
			receipt.Verification.Decision = VerificationFail
			receipt.Verification.ConformanceDecision = VerificationFail
		}
		if generatedResult.Decision == DecisionPass {
			receipt.Summary.PassCount++
		} else if generatedResult.Decision == DecisionFailClosed {
			receipt.Summary.FailClosedCount++
		} else if generatedResult.Decision == DecisionUnknown {
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
	digest, _ := digestJSON(artifact)
	return digest
}

func observationDigest(input Case) string {
	digest, _ := digestJSON(input)
	return digest
}

func claimID(caseID string, rule Rule) string {
	return fmt.Sprintf("gooo://meta-policy-compilation/claim/%s/%02d-%s", caseID, rule.Step, rule.Claim)
}

func claimOutcome(rule Rule, result DecisionResult) (string, string) {
	switch result.Decision {
	case DecisionPass:
		return ClaimDischarged, result.Reason
	case DecisionFailClosed:
		if rule.Step == result.Step {
			return ClaimRefuted, result.Reason
		}
		return ClaimOpen, "FAIL_CLOSED_PENDING_REPAIR"
	default:
		return ClaimOpen, "EVIDENCE_UNAVAILABLE"
	}
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
