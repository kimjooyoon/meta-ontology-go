package publicpartialreuse

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestEvaluateUsesReceiptEvidenceNotScenarioLabels(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "malformed.receipt.json")
	if err := os.WriteFile(filename, append(receiptFixtureBytes(t), '\n', '{', '}'), 0o600); err != nil {
		t.Fatal(err)
	}
	_, contentErr := ReadReceipt(filename)
	if !errors.Is(contentErr, ErrInvalidReceipt) {
		t.Fatalf("native receipt decoder did not identify malformed content: %v", contentErr)
	}
	cases := []struct {
		name     string
		change   func(*EvaluationInput)
		decision string
	}{
		{name: "valid_receipts", decision: DecisionClosed},
		{name: "tampered_label_is_not_evidence", decision: DecisionClosed, change: func(input *EvaluationInput) {
			setReceiptCase(t, input, "TAMPERED_PARTITION_RECEIPT")
		}},
		{name: "missing_receipt", decision: DecisionUnknown, change: func(input *EvaluationInput) {
			delete(input.Receipts, "orders")
		}},
		{name: "unavailable_receipt", decision: DecisionUnknown, change: func(input *EvaluationInput) {
			delete(input.Receipts, "orders")
			input.ReceiptErrors["orders"] = os.ErrNotExist
		}},
		{name: "read_failure_with_previous_object", decision: DecisionUnknown, change: func(input *EvaluationInput) {
			input.ReceiptErrors["orders"] = errors.New("test-only interrupted read")
		}},
		{name: "invalid_content_without_object", decision: DecisionRefuted, change: func(input *EvaluationInput) {
			delete(input.Receipts, "orders")
			input.ReceiptErrors["orders"] = contentErr
		}},
		{name: "invalid_content_with_previous_object", decision: DecisionRefuted, change: func(input *EvaluationInput) {
			input.ReceiptErrors["orders"] = contentErr
		}},
		{name: "refutation_dominates_missing_evidence", decision: DecisionRefuted, change: func(input *EvaluationInput) {
			delete(input.Receipts, "orders")
			delete(input.Receipts, "inventory")
			input.ReceiptErrors["orders"] = contentErr
		}},
		{name: "invalid_in_memory_receipt", decision: DecisionRefuted, change: func(input *EvaluationInput) {
			receipt := input.Receipts["orders"]
			receipt.Original.ExitCode = 1
			input.Receipts["orders"] = receipt
		}},
		{name: "valid_receipt_stale_binding", decision: DecisionUnknown, change: func(input *EvaluationInput) {
			binding := input.Bindings["orders"]
			binding.CanonicalSourceDigest = cache.HashBytes([]byte("changed-test-input")).String()
			input.Bindings["orders"] = binding
		}},
		{name: "ignore_old_receipt_for_executed_partition", decision: DecisionClosed, change: func(input *EvaluationInput) {
			setReceiptCase(t, input, "SINGLE_PARTITION_CHANGE_SELECTIVE_REUSE")
			delete(input.Receipts, "orders")
			input.ReceiptErrors["orders"] = contentErr
		}},
		{name: "reject_invalid_unaffected_partition", decision: DecisionRefuted, change: func(input *EvaluationInput) {
			setReceiptCase(t, input, "SINGLE_PARTITION_CHANGE_SELECTIVE_REUSE")
			delete(input.Receipts, "orders")
			delete(input.Receipts, "inventory")
			input.ReceiptErrors["inventory"] = contentErr
		}},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			input := receiptEvaluationFixture(t)
			if item.change != nil {
				item.change(&input)
			}
			report, err := Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			if err := CompareCase(report, item.decision); err != nil {
				t.Fatal(err)
			}
			if report.Decision != DecisionClosed && (report.ReceiptHits != 0 || report.After.TestUnitsReused != 0) {
				t.Fatalf("non-closed evidence authorized reuse: %+v", report)
			}
		})
	}
}
