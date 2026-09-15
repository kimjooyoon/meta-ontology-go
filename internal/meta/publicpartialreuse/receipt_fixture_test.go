package publicpartialreuse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func receiptEvaluationFixture(t *testing.T) EvaluationInput {
	t.Helper()
	filename := filepath.Join("..", "..", "..", "examples", "self-improvement-partial-reuse", "main.gooo")
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Load(filename, source)
	if err != nil {
		t.Fatal(err)
	}
	item, ok := policy.Case("NO_CHANGE_ALL_RECEIPTS_REUSED")
	if !ok {
		t.Fatal("canonical Gooo policy has no no-change case")
	}
	input := EvaluationInput{
		Policy: policy, Case: item, Bindings: map[string]Binding{},
		Receipts: map[string]Receipt{}, ReceiptErrors: map[string]error{},
	}
	// Synthetic receipt metadata does not attest an actual build or test run.
	digest := cache.HashBytes([]byte("receipt-validation-test-only")).String()
	for _, partition := range policy.Partitions {
		command := TestCommand(partition)
		binding := Binding{
			PolicySourceDigest: policy.SourceDigest, PolicySemanticDigest: policy.SemanticDigest,
			PolicyEvaluatorDigest: policy.EvaluatorDigest, CanonicalSourceDigest: digest,
			CanonicalSemanticSubgraphDigest: digest, GeneratedArtifactDigest: digest,
			GeneratedSemanticDigest: digest, GeneratedManifestDigest: digest,
			CompilerDigest: digest, ReleasedToolDigest: ReleasedToolDigest(digest),
			ToolchainDigest: digest, ToolchainVersion: "go1.27.0", TestCommand: command,
			TestCommandDigest: TestCommandDigest(command, digest), TestContractDigest: digest,
			DependencyGraphDigest: policy.DependencyGraphDigest(), OrchestrationReportDigest: digest,
			OrchestrationOperation: "gooo.self-improvement.public-orchestration",
		}
		receipt := Receipt{
			Schema: ReceiptSchema, Partition: partition.ID, Decision: DecisionClosed,
			Reason: ReceiptReason, Binding: binding,
			Original: OriginalExecution{
				InvocationID: digest, Operation: ReceiptOperation, ResultDigest: digest,
				Successful: true, BuildExecutions: 1, TestExecutions: 1,
				BuildMS: 1, TestMS: 1, WallMS: 1, PeakRSSKib: 1,
			},
			Provenance: Provenance{Operation: Operation, CaseID: item.ID, Stage: "v14-authorized"},
		}
		receipt.ReceiptID, err = ReceiptContentDigest(receipt)
		if err != nil {
			t.Fatal(err)
		}
		input.Bindings[partition.ID], input.Receipts[partition.ID] = binding, receipt
	}
	return input
}

func receiptFixtureBytes(t *testing.T) []byte {
	t.Helper()
	input := receiptEvaluationFixture(t)
	data, err := MarshalReceipt(input.Receipts["orders"])
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func setReceiptCase(t *testing.T, input *EvaluationInput, id string) {
	t.Helper()
	item, ok := input.Policy.Case(id)
	if !ok {
		t.Fatalf("canonical Gooo policy has no case %s", id)
	}
	input.Case = item
	if item.Changed == "orders" {
		input.Execution.TestExecutions = 1
	}
}
