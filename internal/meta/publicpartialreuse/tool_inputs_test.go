package publicpartialreuse

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoooPartialSelectionRejectsChangedExecutionTool(t *testing.T) {
	for _, field := range []string{"path", "executable", "environment", "platform"} {
		t.Run(field, func(t *testing.T) {
			input := receiptEvaluationFixture(t)
			binding := input.Bindings["orders"]
			switch field {
			case "path":
				binding.ToolInputs.Path = "/changed/go"
			case "executable":
				binding.ToolInputs.ExecutableDigest = strings.Repeat("b", 64)
			case "environment":
				binding.ToolInputs.EnvironmentDigest = strings.Repeat("b", 64)
			case "platform":
				binding.ToolInputs.Platform = "changed/host"
			}
			binding.ToolchainDigest = binding.ToolInputs.Digest()
			input.Bindings["orders"] = binding
			selected, err := Select(context.Background(), input)
			if err != nil || selected.Assessment.Decision != DecisionUnknown || len(selected.Partitions) != 0 ||
				selected.Assessment.Reason != "STALE_PARTITION_RECEIPT_BINDING" {
				t.Fatal("changed tool exposed reusable work", selected, err)
			}
			receipt := input.Receipts["orders"]
			receipt.Binding.ToolInputs = binding.ToolInputs
			receipt.ReceiptID, _ = ReceiptContentDigest(receipt)
			if ValidateReceipt(receipt) == nil {
				t.Fatal("resealed contradictory tool observation was accepted")
			}
		})
	}
}

func TestGoooPartialSelectionKeepsHistoricalToolUnknown(t *testing.T) {
	_, err := DecodeReceipt([]byte(`{"schema":"gooo/public-partial-test-reuse-receipt/v1"}`))
	if !errors.Is(err, ErrUnobservedToolInputs) {
		t.Fatal("historical record lost its missing tool observation", err)
	}
	input := receiptEvaluationFixture(t)
	input.ReceiptErrors["orders"] = err
	selected, err := Select(context.Background(), input)
	if err != nil || selected.Assessment.Decision != DecisionUnknown || len(selected.Partitions) != 0 ||
		selected.Assessment.Reason != "PARTITION_EXECUTION_TOOL_UNOBSERVED" || selected.Origin.ActivityID != AssessmentActivityID {
		t.Fatal("Gooo did not own the historical tool decision", selected, err)
	}
	for _, raw := range []string{
		`{"schema":"gooo/public-partial-test-reuse-receipt/v1","unexpected":true}`,
		`{"schema":"gooo/public-partial-test-reuse-receipt/v1"} {}`,
		`{"schema":"wrong","schema":"gooo/public-partial-test-reuse-receipt/v1"}`,
	} {
		if _, err := DecodeReceipt([]byte(raw)); !errors.Is(err, ErrInvalidReceipt) || errors.Is(err, ErrUnobservedToolInputs) {
			t.Fatal("malformed record became a historical tool observation", err)
		}
	}
}

func TestPartialSourceRequiresDeclaredExecutionToolProfile(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "self-improvement-partial-reuse", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range [][2]string{
		{"partial-reuse-schema=v3", "partial-reuse-schema=v2"},
		{"partial-reuse-tool-profile=gooo/generated-test-native-go-command/v1", ""},
		{"partial-reuse-partition-profile=gooo/independent-declaration-partitions/v1", ""},
	} {
		changed := bytes.Replace(source, []byte(edit[0]), []byte(edit[1]), 1)
		if bytes.Equal(changed, source) {
			t.Fatal("source contract marker was not found", edit[0])
		}
		if _, err := Load(path, changed); err == nil || !strings.Contains(err.Error(), "contract v3") {
			t.Fatal("incomplete source contract accepted current tool bindings", err)
		}
	}
}
