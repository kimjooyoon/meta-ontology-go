package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func concurrentScalarSource(index int) []byte {
	return []byte(fmt.Sprintf(`package concurrentdecision%d
namespace concurrentdecision%d
entity Integer id "urn:concurrentdecision%d:integer"
entity Boolean id "urn:concurrentdecision%d:boolean"
activity Balance(Integer) -> Integer id "urn:concurrentdecision%d:balance" computes "return input - %d" assembling {
 choice "direction" operand_order at "0" intent "상수에서 입력을 뺀다. Subtract the input from the constant."
 criteria { predicate "balance" activity "Balanced" input "[-3]" input "[52]" holdout_input "[-11]" }
 attempts "1"
}
activity Balanced(Integer,Integer) -> Boolean computes "return input0 + input1 == %d"
`, index, index, index, index, index, 31+index, 31+index))
}

type concurrentScalarResult struct {
	index  int
	result Result
	err    error
}

func TestScalarSourceModelConcurrentCallsUseSeparateReceipts(t *testing.T) {
	name := scalarBridgeSyntheticArtifact(t)
	g, err := NewTypedPathGenerator(name)
	if err != nil {
		t.Fatal(err)
	}
	info := g.Info()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	responses := make(chan concurrentScalarResult, 4)
	for index := range 4 {
		go func() {
			result, err := g.GenerateSourceAssembly(ctx, fmt.Sprintf("concurrent-%d.gooo", index), concurrentScalarSource(index), "Balance")
			responses <- concurrentScalarResult{index: index, result: result, err: err}
		}()
	}
	verifyConcurrentScalarResponses(t, ctx, responses)
	if !reflect.DeepEqual(g.Info(), info) {
		t.Fatal("concurrent requests changed the retained model information")
	}
	verifyCanceledScalarRequest(t, g)
}

func verifyConcurrentScalarResponses(t *testing.T, ctx context.Context, responses <-chan concurrentScalarResult) {
	t.Helper()
	seen := make(map[string]bool, 4)
	var previous *RecordAssemblyReceipt
	for range 4 {
		select {
		case response := <-responses:
			if response.err != nil {
				t.Fatal(response.err)
			}
			r := response.result.Report.RecordAssembly
			id := fmt.Sprintf("urn:concurrentdecision%d:balance", response.index)
			if r == nil || r.ModelCalls != 1 || r.ScalarPrediction == nil || r.SelectedMask != 1 ||
				len(r.Attempts) != 1 || r.Criteria.Passed != 2 || r.Criteria.Total != 2 ||
				r.CriteriaHoldout.Passed != 1 || r.CriteriaHoldout.Total != 1 || response.result.Report.ActivityID != id {
				t.Fatal("concurrent source obligations or identity differ", r)
			}
			if seen[r.OriginalSourceSHA256] || previous == r || previous != nil && previous.ScalarPrediction == r.ScalarPrediction {
				t.Fatal("concurrent requests shared source identity or prediction state")
			}
			seen[r.OriginalSourceSHA256], previous = true, r
			logConcurrentScalarTiming(t, response.index, r)
		case <-ctx.Done():
			t.Fatal("concurrent source requests did not finish before their deadline", ctx.Err())
		}
	}
}

func logConcurrentScalarTiming(t *testing.T, index int, r *RecordAssemblyReceipt) {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"scope": "one fresh concurrent source request with synthetic untrained weights",
		"index": index, "source_sha256": r.OriginalSourceSHA256, "model_calls": r.ModelCalls,
		"predict_ns": r.PredictNS, "generation_ns": r.GenerationNS,
		"selection_passed": r.Criteria.Passed, "selection_total": r.Criteria.Total,
		"held_passed": r.CriteriaHoldout.Passed, "held_total": r.CriteriaHoldout.Total})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CONCURRENT_SCALAR_OBSERVATION", string(raw))
}

func verifyCanceledScalarRequest(t *testing.T, g *TypedPathGenerator) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := g.GenerateSourceAssembly(ctx, "canceled-concurrent.gooo", concurrentScalarSource(4), "Balance")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("a canceled source request did not retain its cancellation", err)
	}
}
