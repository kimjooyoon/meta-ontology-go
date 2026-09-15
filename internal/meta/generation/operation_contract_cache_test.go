package generation

import (
	"reflect"
	"sync"
	"testing"
)

func TestContractParseReusePreservesValueAndRejectsChangedSource(t *testing.T) {
	var cache operationContractParseCache
	want, err := parseOperationInputContract(operationInputContractSource)
	if err != nil {
		t.Fatal(err)
	}
	first, err := cache.load(operationInputContractSource, [32]byte{})
	if err != nil || !reflect.DeepEqual(first, want) {
		t.Fatalf("cached contract changed: %v", err)
	}
	clear(first.Bindings)
	clear(first.Facts)
	clear(first.ObligationFacts)
	clear(first.PolicyFacts)
	second, err := cache.load(operationInputContractSource, [32]byte{})
	if err != nil || !reflect.DeepEqual(second, want) || cache.parses != 1 {
		t.Fatalf("caller mutation leaked or identical source reparsed: %v parses=%d", err, cache.parses)
	}
	if _, err := cache.load([]byte("invalid contract"), [32]byte{}); err == nil {
		t.Fatal("changed invalid source inherited successful parse")
	}
	if _, err := cache.load(operationInputContractSource, [32]byte{1}); err != nil || cache.parses != 3 {
		t.Fatalf("ABI identity did not force fresh parse: %v parses=%d", err, cache.parses)
	}
}

func TestContractParseReuseSerializesConcurrentReaders(t *testing.T) {
	var cache operationContractParseCache
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			value, err := cache.load(operationInputContractSource, [32]byte{})
			if err != nil || value.SourceDigest == "" || value.SemanticDigest == "" {
				t.Errorf("missing bound contract: %v", err)
			}
			clear(value.PolicyFacts)
		})
	}
	workers.Wait()
	if cache.parses != 1 {
		t.Fatalf("identical concurrent input parsed %d times", cache.parses)
	}
}
