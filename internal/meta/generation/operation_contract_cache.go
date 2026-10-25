package generation

import (
	"crypto/sha256"
	"encoding/json"
	"maps"
	"sync"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

type operationContractCacheKey struct {
	source [32]byte
	abi    [32]byte
}

type operationContractParseCache struct {
	mu     sync.Mutex
	key    operationContractCacheKey
	value  operationInputContract
	valid  bool
	parses uint64
}

var operationContractCache operationContractParseCache

func cachedOperationInputContract(sources []sourcebundle.Source) (operationInputContract, error) {
	abi, err := json.Marshal([]any{nativeOperationInputs, nativeOperationObligations, nativeContractPolicies, nativeContractEntities})
	if err != nil {
		return operationInputContract{}, err
	}
	return operationContractCache.load(sources, sha256.Sum256(abi))
}

// Reuse only parsed embedded-contract data inside this executable. No candidate,
// verifier result, test result, or execution authorization is stored here.
func (cache *operationContractParseCache) load(sources []sourcebundle.Source, abi [32]byte) (operationInputContract, error) {
	digest, err := sourcebundle.Digest(sources)
	if err != nil {
		return operationInputContract{}, err
	}
	key := operationContractCacheKey{source: digest, abi: abi}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.valid && cache.key == key {
		return cloneOperationInputContract(cache.value), nil
	}
	cache.parses++
	value, err := parseOperationInputContractSources(sources)
	if err != nil {
		return operationInputContract{}, err
	}
	cache.key, cache.value, cache.valid = key, value, true
	return cloneOperationInputContract(value), nil
}

func cloneOperationInputContract(value operationInputContract) operationInputContract {
	value.SourceOrigins = value.SourceOrigins.Clone()
	value.Bindings = maps.Clone(value.Bindings)
	value.Facts = maps.Clone(value.Facts)
	value.ObligationFacts = maps.Clone(value.ObligationFacts)
	value.PolicyFacts = maps.Clone(value.PolicyFacts)
	return value
}
