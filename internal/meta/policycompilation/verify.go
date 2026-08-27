package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

func VerifyReceipt(receipt Receipt, policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("unsupported receipt schema %q", receipt.Schema)
	}
	if receipt.Policy.SourceDigest != policy.SourceDigest || receipt.Policy.SemanticDigest != policy.SemanticDigest || receipt.Policy.Denominator != FixedDenominator {
		return errors.New("receipt policy is not bound to the compiled source")
	}
	policyBytes, err := canonicalJSON(policy)
	if err != nil {
		return err
	}
	receiptPolicyBytes, err := canonicalJSON(receipt.Policy)
	if err != nil {
		return err
	}
	if !bytes.Equal(policyBytes, receiptPolicyBytes) {
		return errors.New("receipt policy meaning differs from the compiled policy")
	}
	if receipt.MetaOperation != MetaOperation || receipt.ProofChoice != ProofChoice {
		return errors.New("receipt lost meta-operation or proof choice")
	}
	if receipt.Verification.Decision != VerificationPass || receipt.Verification.ConformanceDecision != VerificationPass || receipt.Verification.SubjectResolution != SubjectUnresolved {
		return errors.New("conformance and subject resolution are not separated")
	}
	if err := verifyWriteSet(receipt.WriteSet); err != nil {
		return err
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return err
	}
	wantProducer := ProducerEvidence{Role: "PRODUCER", Stage: "PRODUCE", Step: 1, Reason: "SOURCE_BOUND", SourceDigest: policy.SourceDigest, SemanticDigest: policy.SemanticDigest, Denominator: policy.Denominator}
	wantConsumer := ConsumerEvidence{Role: "CONSUMER", Stage: "CONSUME", Step: 2, Reason: "ARTIFACT_BOUND", ArtifactSourceDigest: artifact.Policy.SourceDigest, ArtifactDigest: artifactDigest(artifact), SourceMatches: true, RulesMatch: true}
	if receipt.Producer != wantProducer || receipt.Consumer != wantConsumer {
		return errors.New("receipt producer/consumer evidence is not bound to the artifact")
	}
	if receipt.GeneratedDigest != judgeHash || len(receipt.Cases) != len(cases) || len(receipt.Evidence) != len(cases) || receipt.Claims.EventCount != len(receipt.Claims.Events) || len(receipt.Claims.Events) != len(cases)*FixedDenominator*2 {
		return errors.New("receipt denominator or generated artifact binding is incomplete")
	}
	calculatedDigest, err := receiptDigest(receipt)
	if err != nil {
		return err
	}
	if calculatedDigest != receipt.ReceiptDigest {
		return errors.New("receipt digest does not cover the receipt contents")
	}
	if err := verifyClaimLedger(receipt.Claims); err != nil {
		return err
	}
	order := make([]int, len(cases))
	for index := range cases {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return cases[order[i]].ID < cases[order[j]].ID })
	passCount, failClosedCount, unknownCount := 0, 0, 0
	allowedObservations := make(map[string]string, len(cases))
	for _, input := range cases {
		allowedObservations[observationDigest(input)] = input.Provenance
	}
	for _, event := range receipt.Claims.Events {
		if allowedObservations[event.ObservationDigest] != event.Provenance {
			return errors.New("claim transition is not bound to a known case observation")
		}
	}
	for outputIndex, inputIndex := range order {
		input := cases[inputIndex]
		stored := receipt.Cases[outputIndex]
		if stored.ID != input.ID || stored.ValidatorExpectation != input.ValidatorExpectation || stored.EvidenceClass != input.EvidenceClass || stored.Provenance != input.Provenance || stored.ObservationDigest != observationDigest(input) {
			return fmt.Errorf("case %q moved or changed its validator expectation", input.ID)
		}
		if outputIndex == 0 && stored.ClaimStartDigest != "" {
			return fmt.Errorf("case %q has an unexpected claim start", input.ID)
		}
		if outputIndex > 0 && stored.ClaimStartDigest != receipt.Cases[outputIndex-1].ClaimEndDigest {
			return fmt.Errorf("case %q is disconnected from the previous claim segment", input.ID)
		}
		claimEnd := (outputIndex+1)*FixedDenominator*2 - 1
		if claimEnd >= len(receipt.Claims.Events) || stored.ClaimEndDigest != receipt.Claims.Events[claimEnd].Digest {
			return fmt.Errorf("case %q claim segment end is not bound to the ledger", input.ID)
		}
		observed := receipt.Evidence[outputIndex]
		wantObserved := EvidenceObservation{Class: input.EvidenceClass, CaseID: input.ID, ProducerAvailable: input.ProducerAvailable, ConsumerAvailable: input.ConsumerAvailable, SourceDigest: input.ObservedSourceDigest, ArtifactDigest: artifactDigest(artifact), ObservationDigest: observationDigest(input), Provenance: input.Provenance}
		if observed != wantObserved {
			return fmt.Errorf("case %q observation provenance is not bound", input.ID)
		}
		source := EvaluateSourcePolicy(policy, input)
		independent := IndependentEvaluate(policy, input)
		if !sameDecision(stored.Source, source) || !sameDecision(stored.Independent, independent) || !sameDecision(stored.Source, stored.Generated) || !sameDecision(stored.Generated, stored.Independent) || stored.ValidatorExpectation != stored.Source.Decision || !stored.AllDecisionsEquivalent || !stored.DecisionsEquivalent || !stored.ValidatorExpectationConfirmed {
			return fmt.Errorf("case %q is not independently equivalent", input.ID)
		}
		switch stored.Generated.Decision {
		case DecisionPass:
			passCount++
		case DecisionFailClosed:
			failClosedCount++
		case DecisionUnknown:
			unknownCount++
		default:
			return fmt.Errorf("case %q has unsupported decision %q", input.ID, stored.Generated.Decision)
		}
	}
	if receipt.Summary.CaseCount != len(cases) || receipt.Summary.PassCount != passCount || receipt.Summary.FailClosedCount != failClosedCount || receipt.Summary.UnknownCount != unknownCount || receipt.Summary.GeneratedIndependentEqual != len(cases) || receipt.Summary.SourceAllEquivalent != len(cases) || receipt.Summary.ValidatorExpectationsConfirmed != len(cases) {
		return errors.New("case summary does not cover the fixed case denominator")
	}
	if receipt.Verification.Decision != VerificationPass || !receipt.Verification.IndependentReplayed || !receipt.Verification.GeneratedReplayed || !receipt.Verification.LedgerVerified || receipt.Verification.FixedDenominator != FixedDenominator || receipt.Verification.CaseDenominator != len(cases) {
		return errors.New("verification status is not a complete pass")
	}
	return nil
}

func verifyClaimLedger(ledger ClaimLedger) error {
	if ledger.Schema != ClaimLedgerSchema || ledger.EventCount == 0 || ledger.HeadDigest == "" {
		return errors.New("claim ledger header is incomplete")
	}
	prior := ""
	counts := make(map[string]int, len(ledger.Events)/2)
	states := make(map[string]string, len(ledger.Events)/2)
	for index, event := range ledger.Events {
		if event.Event != index+1 || event.PriorDigest != prior || event.ClaimID == "" || event.ObservationDigest == "" || event.Provenance == "" || counts[event.ClaimID] >= 2 {
			return fmt.Errorf("claim ledger chain is broken at event %d", event.Event)
		}
		canonical := event
		canonical.Digest = ""
		digest, err := digestJSON(canonical)
		if err != nil || digest != event.Digest {
			return fmt.Errorf("claim ledger digest is broken at event %d", event.Event)
		}
		if !validClaimTransition(event.From, event.To) {
			return fmt.Errorf("claim ledger transition %q -> %q is invalid", event.From, event.To)
		}
		current, exists := states[event.ClaimID]
		if exists && current != event.From {
			return fmt.Errorf("claim %q does not continue from its previous state", event.ClaimID)
		}
		if !exists && event.From != ClaimUnrecorded {
			return fmt.Errorf("claim %q did not begin at UNRECORDED", event.ClaimID)
		}
		counts[event.ClaimID]++
		states[event.ClaimID] = event.To
		prior = event.Digest
	}
	for claimID, count := range counts {
		if count != 2 || states[claimID] == ClaimUnrecorded {
			return fmt.Errorf("claim %q does not have its persistent two-event lifecycle", claimID)
		}
	}
	if prior != ledger.HeadDigest {
		return errors.New("claim ledger head does not match its event chain")
	}
	return nil
}

func verifyWriteSet(observation WriteSetObservation) error {
	if observation.RepositoryBeforeDigest == "" || observation.RepositoryAfterDigest == "" || observation.RepositoryBeforeDigest != observation.RepositoryAfterDigest || observation.RepositoryWriteChanged || observation.RepositoryBeforeCount != observation.RepositoryAfterCount || observation.GeneratedRootClass != "RUNNER_TEMP_ONLY" || len(observation.GeneratedFiles) != 6 || observation.MutationAuthority != 0 || observation.PromotionAuthority != 0 {
		return errors.New("generated output is not proven runner-temp-only with zero mutation/promotion authority")
	}
	wantFiles := []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"}
	for index, name := range wantFiles {
		if observation.GeneratedFiles[index] != name {
			return errors.New("generated output file set is not canonical")
		}
	}
	return nil
}

func validClaimTransition(from, to string) bool {
	return (from == ClaimUnrecorded && to == ClaimOpen) ||
		(from == ClaimOpen && (to == ClaimOpen || to == ClaimDischarged || to == ClaimRefuted))
}

func DecodeReceipt(data []byte) (Receipt, error) {
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Receipt{}, errors.New("receipt contains trailing JSON")
	}
	return receipt, nil
}
