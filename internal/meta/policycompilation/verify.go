package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

func VerifyReceipt(receipt Receipt, policy CompiledPolicy, artifact PolicyArtifact, judgeHash string, cases []Case) error {
	if receipt.Schema != ReceiptSchema {
		return fmt.Errorf("unsupported receipt schema %q", receipt.Schema)
	}
	if receipt.Policy.SourceDigest != policy.SourceDigest || receipt.Policy.SemanticDigest != policy.SemanticDigest || receipt.Policy.Denominator != FixedDenominator {
		return errors.New("receipt policy is not bound to the compiled source")
	}
	if receipt.MetaOperation != MetaOperation || receipt.ProofChoice != ProofChoice {
		return errors.New("receipt lost meta-operation or proof choice")
	}
	if err := VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return err
	}
	if receipt.GeneratedDigest != judgeHash || len(receipt.Cases) != len(cases) || receipt.Claims.EventCount != len(receipt.Claims.Events) || len(receipt.Claims.Events) != len(cases)*FixedDenominator*2 {
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
	for outputIndex, inputIndex := range order {
		input := cases[inputIndex]
		stored := receipt.Cases[outputIndex]
		if stored.ID != input.ID || stored.Expected != input.Expected {
			return fmt.Errorf("case %q moved or changed its expected decision", input.ID)
		}
		independent := IndependentEvaluate(policy, input)
		if !sameDecision(stored.Independent, independent) || !sameDecision(stored.Generated, stored.Independent) || stored.Expected != stored.Independent.Decision || !stored.DecisionsEquivalent || !stored.ExpectedDecisionConfirmed {
			return fmt.Errorf("case %q is not independently equivalent", input.ID)
		}
	}
	if receipt.Summary.CaseCount != len(cases) || receipt.Summary.GeneratedIndependentEqual != len(cases) || receipt.Summary.ExpectedDecisionsConfirmed != len(cases) {
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
	claims := make(map[string]bool, len(ledger.Events))
	for index, event := range ledger.Events {
		if event.Event != index+1 || event.PriorDigest != prior || event.ClaimID == "" || claims[event.ClaimID+fmt.Sprintf("/%d", event.Event)] {
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
		claims[event.ClaimID+fmt.Sprintf("/%d", event.Event)] = true
		prior = event.Digest
	}
	if prior != ledger.HeadDigest {
		return errors.New("claim ledger head does not match its event chain")
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
	return receipt, nil
}
