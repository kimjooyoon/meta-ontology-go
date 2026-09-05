package generation

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestOperationInputContractSnapshotCacheCallsParserOnceConcurrently(t *testing.T) {
	var parseCalls atomic.Uint64
	cache := newOperationInputContractSnapshotCache(operationInputContractSource, func(raw []byte) (operationInputContract, error) {
		parseCalls.Add(1)
		return parseOperationInputContract(raw)
	})

	const callers = 32
	results := make(chan operationInputContract, callers)
	errors := make(chan error, callers)
	var wait sync.WaitGroup
	wait.Add(callers)
	for range callers {
		go func() {
			defer wait.Done()
			contract, err := cache.load()
			results <- contract
			errors <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errors)

	var first operationInputContract
	for contract := range results {
		if first.SourceDigest == "" {
			first = contract
			continue
		}
		if !reflect.DeepEqual(contract, first) {
			t.Fatalf("cached contract result changed between callers: first=%+v current=%+v", first, contract)
		}
	}
	for err := range errors {
		if err != nil {
			t.Fatalf("cached contract load failed: %v", err)
		}
	}
	if parseCalls.Load() != 1 {
		t.Fatalf("parser invocation count = %d, want 1", parseCalls.Load())
	}
}

func TestOperationInputContractSnapshotCacheRetainsFailure(t *testing.T) {
	invalid := strings.Replace(operationInputContractSource, "activity ExtractFunction(FunctionInput) -> OperationResult\n", "", 1)
	var parseCalls atomic.Uint64
	cache := newOperationInputContractSnapshotCache(invalid, func(raw []byte) (operationInputContract, error) {
		parseCalls.Add(1)
		return parseOperationInputContract(raw)
	})

	first, firstErr := cache.load()
	second, secondErr := cache.load()
	if firstErr == nil || secondErr == nil || firstErr.Error() != secondErr.Error() {
		t.Fatalf("cached failure changed: first=%v second=%v", firstErr, secondErr)
	}
	if !reflect.DeepEqual(first, operationInputContract{}) || !reflect.DeepEqual(second, operationInputContract{}) {
		t.Fatalf("failed cache returned partial contract: first=%+v second=%+v", first, second)
	}
	if parseCalls.Load() != 1 {
		t.Fatalf("failed parser invocation count = %d, want 1", parseCalls.Load())
	}
}

func TestOperationInputContractSnapshotCacheReturnsIndependentMapsAndEvidenceSlice(t *testing.T) {
	cache := newOperationInputContractSnapshotCache(operationInputContractSource, parseOperationInputContract)
	first, err := cache.load()
	if err != nil {
		t.Fatal(err)
	}
	delete(first.Bindings, sourcepolicy.OperationExtractFunction)
	delete(first.Facts, sourcepolicy.OperationExtractFunction)
	delete(first.ObligationFacts, "ProveReturnShape")
	delete(first.PolicyFacts, "NormalizeEligibleImportGroup")

	second, err := cache.load()
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Bindings) != 4 || len(second.Facts) != 4 || len(second.ObligationFacts) != 6 || len(second.PolicyFacts) != 2 {
		t.Fatalf("cached maps were aliased into returned value: %+v", second)
	}

	evidence, err := ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	evidence.Obligations[0].Name = "mutated"
	fresh, err := ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Obligations[0].Name == "mutated" {
		t.Fatal("evidence obligations slice was aliased between calls")
	}
}
