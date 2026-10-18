package publicpartialreuse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const partitionProgramFixture = `package partialreuseexample
type Order struct{}
type Receipt struct{}
type SharedContract struct{}
type StockItem struct{}
type InventorySnapshot struct{}
func CreateReceipt(order Order, shared SharedContract) Receipt { return Receipt{} }
func SnapshotInventory(shared SharedContract, stock StockItem) InventorySnapshot { return InventorySnapshot{} }
`

func revisionFixture(t *testing.T, changed bool) (Policy, []byte) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "examples", "self-improvement-partial-reuse", "main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	program := []byte(partitionProgramFixture)
	if changed {
		for _, pair := range [][2]string{{"entity Receipt id", "entity ReceiptV2 id"}, {"-> Receipt computes", "-> ReceiptV2 computes"},
			{"Order,Receipt,CreateReceipt", "Order,ReceiptV2,CreateReceipt"}} {
			source = bytes.ReplaceAll(source, []byte(pair[0]), []byte(pair[1]))
		}
		program = bytes.ReplaceAll(program, []byte("type Receipt "), []byte("type ReceiptV2 "))
		program = bytes.ReplaceAll(program, []byte(") Receipt { return Receipt{} }"), []byte(") ReceiptV2 { return ReceiptV2{} }"))
	}
	policy, err := Load(path, source)
	if err != nil {
		t.Fatal(err)
	}
	return policy, program
}

// This metadata fixture checks binding rules, not an actual generator run.
func revisionManifest(policy Policy, program []byte) []byte {
	raw, _ := json.Marshal(map[string]any{"schema": "gooo/cli/generate-manifest/v1", "producer": "gooo", "operation": "generate",
		"status": "pass", "source_digest": policy.SourceDigest, "semantic_digest": policy.SemanticDigest,
		"generated_digest": cache.HashBytes(program).String()})
	return raw
}

func revisionBindings(t *testing.T, input EvaluationInput, policy Policy, program []byte) map[string]Binding {
	t.Helper()
	bindings := map[string]Binding{}
	for _, partition := range policy.Partitions {
		old := input.Bindings[partition.ID]
		binding, err := BuildBinding(policy, partition, program, revisionManifest(policy, program), []byte("fixed test contract"), []byte("module fixture\n\ngo 1.27.2\n"),
			old.CompilerDigest, old.OrchestrationReportDigest, old.OrchestrationOperation, old.ToolInputs)
		if err != nil {
			t.Fatal(err)
		}
		bindings[partition.ID] = binding
	}
	return bindings
}

func TestPreviousRevisionReceiptReusesOnlyUnchangedPartition(t *testing.T) {
	input := receiptEvaluationFixture(t)
	beforePolicy, beforeProgram := revisionFixture(t, false)
	afterPolicy, afterProgram := revisionFixture(t, true)
	before := revisionBindings(t, input, beforePolicy, beforeProgram)
	after := revisionBindings(t, input, afterPolicy, afterProgram)
	if before["orders"] == after["orders"] || before["inventory"] != after["inventory"] {
		t.Fatal("partition binding does not retain its own source boundary across revisions")
	}
	for id, receipt := range input.Receipts {
		receipt.Binding = before[id]
		receipt.ReceiptID, _ = ReceiptContentDigest(receipt)
		input.Receipts[id] = receipt
	}
	input.Policy, input.Bindings = afterPolicy, after
	setReceiptCase(t, &input, "SINGLE_PARTITION_CHANGE_SELECTIVE_REUSE")
	selected, err := Select(context.Background(), input)
	if err != nil || selected.Assessment.Decision != DecisionClosed || len(selected.Partitions) != 1 ||
		selected.Partitions[0].ID != "orders" || selected.Assessment.TestUnitsReusable != 1 {
		t.Fatal("Gooo did not select changed work from previous-revision receipts", selected, err)
	}
}

func TestPartitionBindingRetainsSharedInputs(t *testing.T) {
	input := receiptEvaluationFixture(t)
	policy, program := revisionFixture(t, false)
	before := revisionBindings(t, input, policy, program)
	for name, changed := range map[string][]byte{
		"shared-type": bytes.Replace(program, []byte("type SharedContract struct{}"), []byte("type SharedContract struct{ Version int }"), 1),
		"package":     bytes.Replace(program, []byte("package partialreuseexample"), []byte("package changed"), 1),
		"import":      bytes.Replace(program, []byte("type Order"), []byte("import _ \"net/http/pprof\"\ntype Order"), 1),
		"directive":   append([]byte("//go:build custom\n\n"), program...),
	} {
		t.Run(name, func(t *testing.T) {
			after := revisionBindings(t, input, policy, changed)
			for id := range before {
				if before[id] == after[id] {
					t.Fatal("shared input change retained reusable binding", id)
				}
			}
		})
	}
	partition := policy.Partitions[0]
	old := before[partition.ID]
	after, err := BuildBinding(policy, partition, program, revisionManifest(policy, program), []byte("fixed test contract"),
		[]byte("module changed\n\ngo 1.27.2\n"), old.CompilerDigest, old.OrchestrationReportDigest, old.OrchestrationOperation, old.ToolInputs)
	if err != nil || old == after || old.ModuleDigest == after.ModuleDigest {
		t.Fatal("module change retained binding", err)
	}
}

func TestPartitionProfileRejectsUnownedAndCrossPartitionReferences(t *testing.T) {
	policy, program := revisionFixture(t, false)
	for name, declaration := range map[string]string{
		"shared-to-owned": "func SharedRead() Receipt { return Receipt{} }",
		"init-effect":     "func init() { _ = Receipt{} }",
		"cross-owned":     "func CreateReceipt(order Order, shared SharedContract) Receipt { _ = StockItem{}; return Receipt{} }",
	} {
		t.Run(name, func(t *testing.T) {
			changed := append(append([]byte{}, program...), []byte("\n"+declaration)...)
			if _, err := sharedProgramDigest(policy, changed); err == nil {
				t.Fatal("unsupported reference granted an independent partition")
			}
		})
	}
}

func TestPartitionBindingRejectsMismatchedGenerationOrigin(t *testing.T) {
	policy, program := revisionFixture(t, false)
	for _, field := range []string{"schema", "source_digest", "semantic_digest", "generated_digest", "producer", "operation", "status"} {
		t.Run(field, func(t *testing.T) {
			var record map[string]any
			_ = json.Unmarshal(revisionManifest(policy, program), &record)
			record[field] = "different"
			raw, _ := json.Marshal(record)
			if err := validateGenerationOrigin(policy, program, raw); err == nil {
				t.Fatal("generation origin mismatch accepted")
			}
		})
	}
}

func TestGoooRetainsHistoricalPartitionInputUnknown(t *testing.T) {
	_, historical := DecodeReceipt([]byte(`{"schema":"gooo/public-partial-test-reuse-receipt/v2"}`))
	if !errors.Is(historical, ErrUnobservedPartitionInputs) {
		t.Fatal("lost historical input boundary", historical)
	}
	input := receiptEvaluationFixture(t)
	input.ReceiptErrors["orders"] = historical
	selected, err := Select(context.Background(), input)
	if err != nil || selected.Assessment.Decision != DecisionUnknown || len(selected.Partitions) != 0 ||
		selected.Assessment.Reason != "PARTITION_INPUT_SCOPE_UNOBSERVED" {
		t.Fatal("unknown input scope selected work", selected, err)
	}
}
