package semanticdeltareceiptconsumer

import (
	"fmt"
	"os"
	"strings"
)

const (
	TamperInventoryDenominatorID        = "gooo://semantic-delta/tamper-inventory/v1"
	ReplayContextInventoryDenominatorID = "gooo://semantic-delta/replay-context-inventory/v1"
	TamperFixedTotal                    = 12
	ReplayContextFixedTotal             = 4
)

var expectedTamperIDs = []string{
	"proof-choice",
	"stage",
	"step",
	"reason",
	"decision",
	"resolution",
	"classification",
	"expected-subject",
	"observed-checkout",
	"meta-contract",
	"transition-head",
	"effects-status",
}

var expectedReplayContextIDs = []string{
	"exact",
	"parse-unknown",
	"subject-unknown",
	"ambiguous",
}

type ReplayContext struct {
	ID                  string
	BeforePath          string
	AfterPath           string
	RequiresCheckoutSHA bool
}

type TamperContextEvidence struct {
	ID            string `json:"id"`
	RejectedCount int    `json:"rejected_count"`
	FixedTotal    int    `json:"fixed_total"`
	CoverageBPS   int    `json:"coverage_bps"`
}

type TamperMatrixEvidence struct {
	Status               string                  `json:"status"`
	Stage                string                  `json:"stage"`
	Step                 string                  `json:"step"`
	Reason               string                  `json:"reason"`
	DenominatorID        string                  `json:"denominator_id"`
	ExpectedIDs          []string                `json:"expected_ids"`
	ObservedIDs          []string                `json:"observed_ids"`
	RejectedCount        int                     `json:"rejected_count"`
	FixedTotal           int                     `json:"fixed_total"`
	CoverageBPS          int                     `json:"coverage_bps"`
	ContextDenominatorID string                  `json:"context_denominator_id"`
	ExpectedContextIDs   []string                `json:"expected_context_ids"`
	ObservedContextIDs   []string                `json:"observed_context_ids"`
	ContextFixedTotal    int                     `json:"context_fixed_total"`
	ContextCoverageBPS   int                     `json:"context_coverage_bps"`
	Contexts             []TamperContextEvidence `json:"contexts"`
}

type tamperMutation struct {
	id   string
	edit func(*Receipt)
}

func FixedTamperIDs() []string {
	return append([]string(nil), expectedTamperIDs...)
}

func FixedReplayContextIDs() []string {
	return append([]string(nil), expectedReplayContextIDs...)
}

func FixedReplayContexts() []ReplayContext {
	return []ReplayContext{
		{ID: "exact", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/equivalent-after.gooo", RequiresCheckoutSHA: true},
		{ID: "parse-unknown", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/indeterminate-after.gooo", RequiresCheckoutSHA: true},
		{ID: "subject-unknown", BeforePath: "examples/semantic-delta-receipt/before.gooo", AfterPath: "examples/semantic-delta-receipt/equivalent-after.gooo", RequiresCheckoutSHA: false},
		{ID: "ambiguous", BeforePath: "examples/semantic-delta-receipt/ambiguous-before.gooo", AfterPath: "examples/semantic-delta-receipt/ambiguous-after.gooo", RequiresCheckoutSHA: true},
	}
}

func BuildTamperMatrix(subjectSHA, observedCheckoutSHA string) (TamperMatrixEvidence, error) {
	mutations := fixedTamperMutations()
	observedTamperIDs := make([]string, 0, len(mutations))
	for _, mutation := range mutations {
		observedTamperIDs = append(observedTamperIDs, mutation.id)
	}
	contexts := FixedReplayContexts()
	observedContextIDs := make([]string, 0, len(contexts))
	for _, context := range contexts {
		observedContextIDs = append(observedContextIDs, context.ID)
	}
	evidence := TamperMatrixEvidence{
		Status:               "FAIL_CLOSED",
		Stage:                "tamper-inventory",
		Step:                 "validate-ids",
		Reason:               "TAMPER_INVENTORY_MISMATCH",
		DenominatorID:        TamperInventoryDenominatorID,
		ExpectedIDs:          FixedTamperIDs(),
		ObservedIDs:          observedTamperIDs,
		FixedTotal:           TamperFixedTotal,
		ContextDenominatorID: ReplayContextInventoryDenominatorID,
		ExpectedContextIDs:   FixedReplayContextIDs(),
		ObservedContextIDs:   observedContextIDs,
		ContextFixedTotal:    ReplayContextFixedTotal,
	}
	if err := exactIDInventory(evidence.ExpectedIDs, evidence.ObservedIDs); err != nil {
		return evidence, fmt.Errorf("tamper inventory: %w", err)
	}
	if err := exactIDInventory(evidence.ExpectedContextIDs, evidence.ObservedContextIDs); err != nil {
		return evidence, fmt.Errorf("replay context inventory: %w", err)
	}

	minimumRejected := TamperFixedTotal
	passingContexts := 0
	for _, context := range contexts {
		contextSHA := observedCheckoutSHA
		if !context.RequiresCheckoutSHA {
			contextSHA = ""
		}
		input := Input{CaseID: context.ID, SubjectSHA: subjectSHA, ObservedCheckoutSHA: contextSHA, BeforePath: context.BeforePath, AfterPath: context.AfterPath}
		beforeRaw, err := os.ReadFile(input.BeforePath)
		if err != nil {
			return evidence, err
		}
		afterRaw, err := os.ReadFile(input.AfterPath)
		if err != nil {
			return evidence, err
		}
		expected := reconstructReceipt(input, beforeRaw, afterRaw)
		rejected := 0
		for _, mutation := range mutations {
			tampered := expected
			mutation.edit(&tampered)
			tampered.ReceiptDigest = ""
			tampered.ReceiptDigest = digestValue(tampered)
			if !AdjudicateFiles(input, tampered).Passed {
				rejected++
			}
		}
		if rejected < minimumRejected {
			minimumRejected = rejected
		}
		if rejected == TamperFixedTotal {
			passingContexts++
		}
		evidence.Contexts = append(evidence.Contexts, TamperContextEvidence{ID: context.ID, RejectedCount: rejected, FixedTotal: TamperFixedTotal, CoverageBPS: coverageBPS(rejected, TamperFixedTotal)})
	}
	evidence.RejectedCount = minimumRejected
	evidence.CoverageBPS = coverageBPS(evidence.RejectedCount, evidence.FixedTotal)
	evidence.ContextCoverageBPS = coverageBPS(passingContexts, evidence.ContextFixedTotal)
	if passingContexts == ReplayContextFixedTotal && evidence.RejectedCount == TamperFixedTotal {
		evidence.Status, evidence.Stage, evidence.Step, evidence.Reason = "EXACT", "tamper-matrix", "reject-resealed", "ALL_FIXED_TAMPERS_REJECTED"
		return evidence, nil
	}
	evidence.Stage, evidence.Step, evidence.Reason = "tamper-matrix", "reject-resealed", "TAMPER_REJECTION_INCOMPLETE"
	return evidence, fmt.Errorf("tamper matrix rejected %d/%d across %d/%d contexts", evidence.RejectedCount, evidence.FixedTotal, passingContexts, evidence.ContextFixedTotal)
}

func exactIDInventory(expected, observed []string) error {
	if len(expected) == 0 || len(expected) != len(observed) {
		return fmt.Errorf("fixed inventory cardinality mismatch: expected=%d observed=%d", len(expected), len(observed))
	}
	counts := make(map[string]int, len(observed))
	for _, id := range observed {
		counts[id]++
	}
	for _, id := range expected {
		if counts[id] != 1 {
			return fmt.Errorf("fixed inventory ID mismatch at %q", id)
		}
		counts[id] = 0
	}
	for id, count := range counts {
		if count != 0 {
			return fmt.Errorf("unexpected fixed inventory ID %q", id)
		}
	}
	return nil
}

func fixedTamperMutations() []tamperMutation {
	return []tamperMutation{
		{id: "proof-choice", edit: func(r *Receipt) { r.ProofChoice = "TAMPERED" }},
		{id: "stage", edit: func(r *Receipt) { r.Stage = "TAMPERED" }},
		{id: "step", edit: func(r *Receipt) { r.Step = "TAMPERED" }},
		{id: "reason", edit: func(r *Receipt) { r.Reason = "TAMPERED" }},
		{id: "decision", edit: func(r *Receipt) { r.Decision = "TAMPERED" }},
		{id: "resolution", edit: func(r *Receipt) { r.Resolution = "TAMPERED" }},
		{id: "classification", edit: func(r *Receipt) { r.Classification = "TAMPERED" }},
		{id: "expected-subject", edit: func(r *Receipt) { r.ExpectedSubjectSHA = strings.Repeat("b", 40) }},
		{id: "observed-checkout", edit: func(r *Receipt) { r.ObservedCheckoutSHA = strings.Repeat("b", 40) }},
		{id: "meta-contract", edit: func(r *Receipt) { r.MetaContractDigest = "sha256:" + strings.Repeat("b", 64) }},
		{id: "transition-head", edit: func(r *Receipt) { r.TransitionHeadDigest = "sha256:" + strings.Repeat("b", 64) }},
		{id: "effects-status", edit: func(r *Receipt) { r.Effects.Status = "NET_REPOSITORY_STATE_CHANGED" }},
	}
}
