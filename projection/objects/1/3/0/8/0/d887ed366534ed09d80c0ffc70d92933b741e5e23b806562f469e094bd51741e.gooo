package workfrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func bindR4Rules(t *testing.T, input R4Input, name string) R4Input {
	t.Helper()
	if name == "missing-bound" {
		return input
	}
	graph, err := AnalyzeR4Graph(input)
	if err != nil {
		t.Fatal(err)
	}
	var digest string
	for _, component := range graph.SCCs {
		if component.Cyclic {
			digest = component.Digest
		}
	}
	if name == "stale-digest" {
		digest = "stale-scc-digest"
	}
	maxIterations := uint64(2)
	iterationsUsed := uint64(0)
	if name == "zero-bound" {
		maxIterations = 0
	}
	if name == "iteration-exhaustion" {
		maxIterations = 1
		iterationsUsed = 1
	}
	input.Rules = []R4Rule{{SCCDigest: digest, MaxIterations: maxIterations, IterationsUsed: iterationsUsed}}
	if name == "conflicting-rule" {
		input.Rules = append(input.Rules, R4Rule{SCCDigest: digest, MaxIterations: maxIterations + 1})
	}
	return input
}
func r4BindingDigest(payload string) string {
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}
func r4FixtureDigest(input R4Input) string {
	data, err := EncodeR4JSON(input)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
func reversePressures(values []Pressure) []Pressure {
	result := append([]Pressure(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
func reverseStates(values []ObligationState) []ObligationState {
	result := append([]ObligationState(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
func reversePaths(values []RepairPath) []RepairPath {
	result := append([]RepairPath(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
