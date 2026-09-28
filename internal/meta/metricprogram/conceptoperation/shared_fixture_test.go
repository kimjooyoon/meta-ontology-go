package conceptoperation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

type receiptFixtureSnapshot struct {
	Receipt Receipt
	Inputs  SourceInputs
}

var receiptFixtureOnce sync.Once
var receiptFixtureBase receiptFixture
var receiptFixtureBytes []byte
var receiptFixtureDirectory string
var receiptFixturePreparations atomic.Uint64
var receiptFixtureCopies atomic.Uint64

// Only producer input data is shared. Every consumer verification still runs.
func buildReceiptFixture(t *testing.T) receiptFixture {
	t.Helper()
	receiptFixtureOnce.Do(func() {
		directory, err := os.MkdirTemp("", "gooo-concept-operation-fixture-")
		if err != nil {
			t.Fatal(err)
		}
		receiptFixtureDirectory = directory
		receiptFixturePreparations.Add(1)
		receiptFixtureBase = buildFreshReceiptFixture(t, directory)
		payload, err := json.Marshal(receiptFixtureSnapshot{
			Receipt: receiptFixtureBase.receipt,
			Inputs:  receiptFixtureBase.inputs,
		})
		if err != nil {
			t.Fatal(err)
		}
		receiptFixtureBytes = payload
	})
	if len(receiptFixtureBytes) == 0 {
		t.Fatal("shared producer fixture initialization did not complete")
	}
	var snapshot receiptFixtureSnapshot
	if err := json.Unmarshal(receiptFixtureBytes, &snapshot); err != nil {
		t.Fatal(err)
	}
	fixture := receiptFixtureBase
	fixture.receipt, fixture.inputs = snapshot.Receipt, snapshot.Inputs
	fixture.inputs.ScratchDirectory = t.TempDir()
	receiptFixtureCopies.Add(1)
	return fixture
}

// The shared producer directory must outlive the first test's Cleanup calls.
func TestMain(m *testing.M) {
	code := m.Run()
	if receiptFixtureDirectory != "" {
		if err := os.RemoveAll(receiptFixtureDirectory); err != nil {
			fmt.Fprintln(os.Stderr, err)
			code = 1
		}
	}
	event := make(map[string]any)
	event["reuse_scope"] = "TEST_PRODUCER_INPUT_SNAPSHOT_ONLY"
	event["producer_preparations_started"] = receiptFixturePreparations.Load()
	event["producer_prepared"] = len(receiptFixtureBytes) != 0
	event["fixture_copies"] = receiptFixtureCopies.Load()
	event["snapshot_digest"] = fixtureDigest(string(receiptFixtureBytes))
	event["test_exit_code"] = code
	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	} else {
		fmt.Fprintf(os.Stdout, "CONCEPT_OPERATION_FIXTURE_PREPARATION=%s\n", payload)
	}
	os.Exit(code)
}

func TestReceiptFixtureCopiesAreIndependent(t *testing.T) {
	first := buildReceiptFixture(t)
	originalReceipt, err := json.Marshal(first.receipt)
	if err != nil {
		t.Fatal(err)
	}
	inputs := receiptFixturePayloads(first.inputs)
	originalInputs := make([][]byte, 0, len(inputs))
	for _, payload := range inputs {
		if len(payload) == 0 {
			t.Fatal("producer fixture omitted a required payload")
		}
		originalInputs = append(originalInputs, bytes.Clone(payload))
		payload[0] ^= 1
	}
	first.receipt.Expected[0].Activity = "caller-local-mutation"
	second := buildReceiptFixture(t)
	secondReceipt, err := json.Marshal(second.receipt)
	if err != nil || !bytes.Equal(originalReceipt, secondReceipt) {
		t.Fatalf("caller receipt mutation reached shared producer data: %v", err)
	}
	for index, payload := range receiptFixturePayloads(second.inputs) {
		if !bytes.Equal(originalInputs[index], payload) {
			t.Fatalf("caller payload %d mutation reached another fixture", index)
		}
	}
	if first.inputs.ScratchDirectory == second.inputs.ScratchDirectory ||
		first.inputs.ScratchDirectory == receiptFixtureDirectory ||
		second.inputs.ScratchDirectory == receiptFixtureDirectory {
		t.Fatal("caller scratch directories share producer storage")
	}
	if receiptFixturePreparations.Load() != 1 {
		t.Fatal("immutable producer fixture was prepared more than once")
	}
}

func receiptFixturePayloads(inputs SourceInputs) [][]byte {
	return [][]byte{
		inputs.Strategy, inputs.StrategyVerification,
		inputs.SourceMetrics, inputs.Intervention,
		inputs.InterventionVerification, inputs.Program,
		inputs.ProgramSource, inputs.ProgramVerification,
	}
}
