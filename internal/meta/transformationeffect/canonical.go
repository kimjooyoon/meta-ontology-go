package transformationeffect

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func hashBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func hashJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return hashBytes(payload)
}

func hashPair(left, right string) string { return hashBytes([]byte(left + "\x00" + right)) }

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sealLedger(ledger Ledger) Ledger {
	ledger.EffectDigest = hashJSON(ledger.Effects)
	ledger.SemanticDigest = hashJSON([]any{ledger.HeadSHA, ledger.Decision, ledger.IndicatorLedgerDigest,
		ledger.SourceTreeBefore, ledger.SandboxTreeAfter, ledger.EffectDigest, ledger.PatchDigest,
		ledger.GeneratedReceiptReportDigest, ledger.ExecutedProvenanceDigest, ledger.Indicators})
	ledger.LedgerDigest, ledger.ReplayDigest = "", ""
	ledger.LedgerDigest = hashJSON(ledger)
	ledger.ReplayDigest = hashPair(ledger.InputDigest, ledger.LedgerDigest)
	return ledger
}

func validateLedger(ledger Ledger) error {
	if ledger.Schema != ledgerSchema || ledger.Status != "BOUND" || !ledger.RootTopologyExempt ||
		ledger.PromotionAuthorized || !validSHA(ledger.BaseSHA) || !validSHA(ledger.HeadSHA) ||
		ledger.SelectedPlanOperations < 0 || ledger.BoundExecutorOperations != ledger.SelectedPlanOperations ||
		ledger.UnboundExecutorOperations != 0 || len(ledger.Effects) != ledger.SelectedPlanOperations {
		return fmt.Errorf("transformation ledger is not bound")
	}
	if ledger.ReceiptCount < 0 || ledger.FailureCount < 0 || ledger.UnknownCount < 0 ||
		ledger.ReceiptCount+ledger.FailureCount != ledger.SelectedPlanOperations ||
		ledger.ReceiptDecision == "" || ledger.OperationOutcome == "" {
		return fmt.Errorf("transformation ledger operation outcome is incomplete")
	}
	if ledger.Decision == "FIXED_POINT" {
		if ledger.OperationOutcome != OperationOutcomeFixedPoint ||
			ledger.ReceiptDecision != string(generation.ReceiptDecisionFixedPoint) ||
			ledger.ReceiptCount != 0 || ledger.FailureCount != 0 {
			return fmt.Errorf("fixed point operation outcome is inconsistent")
		}
	}
	if ledger.Decision == "APPLIED" && ledger.OperationOutcome != OperationOutcomeClosed &&
		ledger.OperationOutcome != OperationOutcomeMixedClosedRefuted &&
		ledger.OperationOutcome != OperationOutcomeRefuted {
		return fmt.Errorf("applied operation outcome is inconsistent")
	}
	for _, indicator := range ledger.Indicators {
		if indicator.Verdict != "PASS" {
			return fmt.Errorf("indicator %s did not pass", indicator.ID)
		}
	}
	expected := ledger
	expected.EffectDigest, expected.SemanticDigest, expected.LedgerDigest, expected.ReplayDigest = "", "", "", ""
	if !reflect.DeepEqual(sealLedger(expected), ledger) {
		return fmt.Errorf("transformation ledger digest diverged")
	}
	return nil
}

func encodeJSON(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	return append(payload, '\n'), err
}
