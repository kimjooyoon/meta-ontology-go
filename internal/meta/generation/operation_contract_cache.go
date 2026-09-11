package generation

import (
	"crypto/sha256"
	"encoding/json"
	"maps"
	"sync"
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

func cachedOperationInputContract(raw []byte) (operationInputContract, error) {
	abi, err := json.Marshal([]any{nativeOperationInputs, nativeOperationObligations, nativeContractPolicies})
	if err != nil {
		return operationInputContract{}, err
	}
	return operationContractCache.load(raw, sha256.Sum256(abi))
}

// Reuse only parsed embedded-contract data inside this executable. No candidate,
// verifier result, test result, or execution authorization is stored here.
func (cache *operationContractParseCache) load(raw []byte, abi [32]byte) (operationInputContract, error) {
	key := operationContractCacheKey{source: sha256.Sum256(raw), abi: abi}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.valid && cache.key == key {
		return cloneOperationInputContract(cache.value), nil
	}
	cache.parses++
	value, err := parseOperationInputContract(raw)
	if err != nil {
		return operationInputContract{}, err
	}
	cache.key, cache.value, cache.valid = key, value, true
	return cloneOperationInputContract(value), nil
}

func cloneOperationInputContract(value operationInputContract) operationInputContract {
	value.Bindings = maps.Clone(value.Bindings)
	value.Facts = maps.Clone(value.Facts)
	value.ObligationFacts = maps.Clone(value.ObligationFacts)
	value.PolicyFacts = maps.Clone(value.PolicyFacts)
	return value
}
