package main

import (
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type semanticReuseCache struct {
	file           *syntax.File
	previousDigest cache.Digest
	inputDigest    cache.Digest
	result         generationResult
}

func generateWithDeadlineAdopted(file *syntax.File, previous []byte, timeout time.Duration, authorization generation.SemanticAdoptionAuthorization, reuse *semanticReuseCache) (generationResult, bool, error) {
	previousDigest := cache.HashBytes(previous)
	inputDigest := cache.Digest(authorization.CandidateInputDigest)
	if reuse != nil && reuse.file == file && reuse.previousDigest == previousDigest && reuse.inputDigest == inputDigest {
		return reuse.result, true, nil
	}
	result, err := generateWithDeadlineCore(file, previous, timeout)
	if err != nil {
		return result, false, err
	}
	digest, err := cache.SemanticDigest(result.ir)
	if err != nil {
		return generationResult{}, false, fmt.Errorf("adopted semantic digest failed: %w", err)
	}
	if digest != inputDigest {
		return generationResult{}, false, fmt.Errorf("authorized candidate input digest %s does not match compiler digest %s", inputDigest, digest)
	}
	if reuse != nil {
		*reuse = semanticReuseCache{file: file, previousDigest: previousDigest, inputDigest: inputDigest, result: result}
	}
	return result, false, nil
}
