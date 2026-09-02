package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

func VerifyCompiledArtifact(artifact PolicyArtifact, policy CompiledPolicy, judgeHash string) error {
	if artifact.Schema != ArtifactSchema || artifact.GeneratedJudgeHash != judgeHash {
		return errors.New("compiled artifact header is not bound to the generated judge")
	}
	want, err := canonicalJSON(policy)
	if err != nil {
		return err
	}
	got, err := canonicalJSON(artifact.Policy)
	if err != nil {
		return err
	}
	if !bytes.Equal(want, got) || artifact.Policy.SourceDigest != policy.SourceDigest || artifact.Policy.SemanticDigest != policy.SemanticDigest || artifact.Policy.Denominator != FixedDenominator {
		return errors.New("compiled artifact policy meaning differs from source policy")
	}
	return nil
}

func VerifyReceipt(receipt Receipt, policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case) error {
	if receipt.Schema != ReceiptSchema || len(cases) != ExpectedCaseCount {
		return errors.New("receipt schema or canonical case denominator is invalid")
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
		return errors.New("receipt policy meaning differs from compiled policy")
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return err
	}
	if receipt.MetaOperation == "" || receipt.ProofChoice == "" || receipt.GeneratedDigest != judgeHash {
		return errors.New("receipt source bindings are incomplete")
	}
	if receipt.Verification.Decision != VerificationPass || receipt.Verification.ConformanceDecision != VerificationPass || receipt.Verification.FixedDenominator != FixedDenominator || receipt.Verification.CaseDenominator != ExpectedCaseCount {
		return errors.New("receipt verification status is not a complete pass")
	}
	if receipt.CurrentEvidence.Class != EvidenceCurrent || receipt.CurrentEvidence.SourceDigest != policy.SourceDigest || receipt.CurrentEvidence.ArtifactSourceDigest != artifact.Policy.SourceDigest || receipt.CurrentEvidence.GeneratedJudgeDigest != judgeHash || receipt.CurrentEvidence.IndependentDigest != policy.SemanticDigest || receipt.CurrentEvidence.Provenance == "" {
		return errors.New("current evidence is missing or mixed with synthetic evidence")
	}
	if receipt.PublicCLI.Path != "gooo" || receipt.PublicCLI.CheckExit != 0 || receipt.PublicCLI.GenerateExit != 0 || !receipt.PublicCLI.CheckObserved || !receipt.PublicCLI.GenerateObserved {
		return errors.New("public Gooo CLI evidence is incomplete")
	}
	if len(receipt.Evidence) != ExpectedCaseCount || receipt.Claims.EventCount != len(receipt.Claims.Events) || len(receipt.Claims.Events) != ExpectedCaseCount*ClaimPredicateCount*2 {
		return errors.New("receipt evidence or claim denominator is incomplete")
	}
	calculated, err := ReceiptDigest(receipt)
	if err != nil {
		return err
	}
	if calculated != receipt.ReceiptDigest {
		return errors.New("receipt digest does not cover receipt contents")
	}
	if err := verifyWriteSet(receipt.WriteSet); err != nil {
		return err
	}
	if err := verifyClaimLedger(receipt.Claims); err != nil {
		return err
	}
	if err := verifyClaimPredicates(receipt, policy); err != nil {
		return err
	}

	byID := make(map[string]Case, len(cases))
	for _, input := range cases {
		if input.ID == "" || byID[input.ID].ID != "" {
			return errors.New("case IDs must be unique")
		}
		byID[input.ID] = input
	}
	for index, stored := range receipt.Cases {
		input, ok := byID[stored.ID]
		if !ok || stored.ValidatorExpectation != input.ValidatorExpectation || stored.EvidenceClass != input.EvidenceClass || stored.Provenance != input.Provenance || stored.ObservationDigest != ObservationDigest(input) {
			return fmt.Errorf("case %q is not bound to the canonical input", stored.ID)
		}
		if index > 0 && stored.ClaimStartDigest != receipt.Cases[index-1].ClaimEndDigest {
			return fmt.Errorf("case %q claim segment is disconnected", stored.ID)
		}
		end := (index+1)*ClaimPredicateCount*2 - 1
		if stored.ClaimEndDigest != receipt.Claims.Events[end].Digest {
			return fmt.Errorf("case %q claim segment is not bound to ledger", stored.ID)
		}
		if receipt.Evidence[index] != evidenceForCase(input, policy) {
			return fmt.Errorf("case %q synthetic evidence is not bound", stored.ID)
		}
		source := EvaluateSourcePolicy(policy, input)
		independent := IndependentEvaluate(policy, input)
		if !sameResult(stored.Source, source) || !sameResult(stored.Generated, stored.Source) || !sameResult(stored.Independent, independent) || !sameResult(stored.Independent, stored.Source) || !stored.AllDecisionsEquivalent || !stored.DecisionsEquivalent || !stored.ValidatorExpectationConfirmed {
			return fmt.Errorf("case %q source/generated/independent results diverge", stored.ID)
		}
		if err := verifyCaseClaimEvents(receipt.Claims, index, stored, policy); err != nil {
			return err
		}
	}
	if receipt.Summary.CaseCount != ExpectedCaseCount || receipt.Summary.CanonicalCaseDenominator != ExpectedCaseCount || receipt.Summary.GeneratedIndependentEquivalent != ExpectedCaseCount || receipt.Summary.SourceAllEquivalent != ExpectedCaseCount || receipt.Summary.ValidatorExpectationsConfirmed != ExpectedCaseCount {
		return errors.New("receipt summary does not cover 3/3/3 canonical cases")
	}
	return nil
}

func verifyWriteSet(observation WriteSetObservation) error {
	if observation.RepositoryBeforeDigest == "" || observation.RepositoryAfterDigest == "" || observation.RepositoryBeforeDigest != observation.RepositoryAfterDigest || observation.RepositoryNetChangeObserved || observation.RepositoryBeforeCount != observation.RepositoryAfterCount || observation.GeneratedRootClass != "RUNNER_TEMP_ONLY" || observation.MutationAuthority != 0 || observation.PromotionAuthority != 0 {
		return errors.New("repository net-write claim is not an unchanged exact boundary")
	}
	want := []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"}
	if len(observation.GeneratedFiles) != len(want) {
		return errors.New("generated output file denominator changed")
	}
	for index := range want {
		if observation.GeneratedFiles[index] != want[index] {
			return errors.New("generated output file set is not canonical")
		}
	}
	return nil
}

func verifyClaimLedger(ledger ClaimLedger) error {
	if ledger.Schema != ClaimLedgerSchema || ledger.EventCount == 0 || ledger.HeadDigest == "" {
		return errors.New("claim ledger header is incomplete")
	}
	prior := ""
	counts := make(map[string]int)
	states := make(map[string]string)
	for index, event := range ledger.Events {
		if event.Event != index+1 || event.PriorDigest != prior || event.ClaimID == "" || event.Predicate == "" || event.ObservationDigest == "" || event.Provenance == "" || counts[event.ClaimID] >= 2 {
			return fmt.Errorf("claim ledger chain is broken at event %d", event.Event)
		}
		canonical := event
		canonical.Digest = ""
		want, err := digestJSON(canonical)
		if err != nil || want != event.Digest {
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
		if event.From == ClaimUnrecorded && (event.To != ClaimOpen || event.Observed) {
			return fmt.Errorf("claim %q has an invalid opening transition", event.ClaimID)
		}
		if event.From == ClaimOpen && ((event.To == ClaimDischarged || event.To == ClaimRefuted) != event.Observed) {
			return fmt.Errorf("claim %q has an invalid outcome transition", event.ClaimID)
		}
		counts[event.ClaimID]++
		states[event.ClaimID] = event.To
		prior = event.Digest
	}
	if prior != ledger.HeadDigest {
		return errors.New("claim ledger head does not match event chain")
	}
	for claimID, count := range counts {
		if count != 2 || states[claimID] == ClaimUnrecorded {
			return fmt.Errorf("claim %q does not have its two-event lifecycle", claimID)
		}
	}
	return nil
}

func validClaimTransition(from, to string) bool {
	return (from == ClaimUnrecorded && to == ClaimOpen) || (from == ClaimOpen && (to == ClaimOpen || to == ClaimDischarged || to == ClaimRefuted))
}

func verifyClaimPredicates(receipt Receipt, policy CompiledPolicy) error {
	if len(receipt.Cases) != ExpectedCaseCount {
		return errors.New("claim predicate case denominator is not 3")
	}
	for _, stored := range receipt.Cases {
		if len(stored.ClaimPredicates) != ClaimPredicateCount {
			return fmt.Errorf("case %q does not have eight claim predicates", stored.ID)
		}
		seen := make(map[string]bool, ClaimPredicateCount)
		for _, predicate := range stored.ClaimPredicates {
			if seen[predicate.Predicate] || predicate.ClaimID == "" || predicate.ObservationDigest != stored.ObservationDigest || predicate.Provenance != stored.Provenance || (predicate.Outcome != ClaimOpen && predicate.Outcome != ClaimDischarged && predicate.Outcome != ClaimRefuted) {
				return fmt.Errorf("case %q has a duplicate or malformed predicate", stored.ID)
			}
			seen[predicate.Predicate] = true
		}
		for _, rule := range policy.Rules {
			if !seen[rule.Claim] {
				return fmt.Errorf("case %q is missing source predicate %q", stored.ID, rule.Claim)
			}
		}
	}
	return nil
}

func verifyCaseClaimEvents(ledger ClaimLedger, caseIndex int, stored CaseReceipt, policy CompiledPolicy) error {
	base := caseIndex * ClaimPredicateCount * 2
	byID := make(map[string]ClaimPredicateObservation, ClaimPredicateCount)
	for _, observation := range stored.ClaimPredicates {
		byID[observation.ClaimID] = observation
	}
	for ruleIndex, rule := range policy.Rules {
		id := claimID(stored.ID, rule)
		observation, ok := byID[id]
		if !ok {
			return fmt.Errorf("case %q has no observation for predicate %q", stored.ID, rule.Claim)
		}
		opening := ledger.Events[base+ruleIndex]
		outcome := ledger.Events[base+ClaimPredicateCount+ruleIndex]
		if opening.ClaimID != id || opening.Predicate != rule.Claim || opening.From != ClaimUnrecorded || opening.To != ClaimOpen || opening.Decision != stored.Generated.Decision || opening.Stage != rule.Stage || opening.Step != rule.Step || outcome.ClaimID != id || outcome.Predicate != rule.Claim || outcome.From != ClaimOpen || outcome.To != observation.Outcome || outcome.Decision != stored.Generated.Decision || outcome.Stage != observation.Stage || outcome.Step != observation.Step || outcome.Reason != observation.Reason || outcome.Observed != observation.Observed {
			return fmt.Errorf("case %q predicate %q is not independently ledger-bound", stored.ID, rule.Claim)
		}
	}
	return nil
}

func DecodeReceipt(data []byte) (Receipt, error) {
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return Receipt{}, errors.New("receipt contains trailing JSON")
	}
	return receipt, nil
}
