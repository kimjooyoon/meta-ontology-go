package workfrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func r4ProjectionBytes(input R4Input) (r4ProjectionSet, error) {
	input = normalizeR4Input(input)
	iterationUses := make([]r4IterationUse, 0, len(input.Rules))
	maxIterations := make([]r4MaxIterationPolicy, 0, len(input.Rules))
	for _, rule := range input.Rules {
		iterationUses = append(iterationUses, r4IterationUse{SCCDigest: rule.SCCDigest, IterationsUsed: rule.IterationsUsed})
		maxIterations = append(maxIterations, r4MaxIterationPolicy{SCCDigest: rule.SCCDigest, MaxIterations: rule.MaxIterations})
	}
	values := []any{
		r4SnapshotProjection{
			SchemaVersion: R4SchemaVersion, Roots: input.RootObligationIDs,
			States: input.States, Paths: input.Paths, IterationUses: iterationUses,
		},
		r4PolicyProjection{
			SchemaVersion: R4SchemaVersion, MinimumSelectedPressures: input.MinimumSelectedPressures,
			Capacity: input.Capacity, MaxIterations: maxIterations,
		},
		r4RegistryProjection{SchemaVersion: R4SchemaVersion, Pressures: input.Pressures},
	}
	encoded := make([][]byte, len(values))
	for index, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return r4ProjectionSet{}, fmt.Errorf("encode r4 projection: %w", err)
		}
		encoded[index] = data
	}
	return r4ProjectionSet{snapshot: encoded[0], policy: encoded[1], registry: encoded[2]}, nil
}
func r4PayloadBinding(payload []byte) (string, string) {
	digest := sha256.Sum256(payload)
	return string(payload), hex.EncodeToString(digest[:])
}
func validateR4Bindings(input R4Input) string {
	projections, err := r4ProjectionBytes(input)
	if err != nil {
		return R4ReasonMalformedBinding
	}
	bindings := []r4Binding{
		{payload: input.SnapshotPayload, digest: input.SnapshotDigest, reason: R4ReasonSnapshotBindingMismatch, expected: projections.snapshot},
		{payload: input.PolicyPayload, digest: input.PolicyDigest, reason: R4ReasonPolicyBindingMismatch, expected: projections.policy},
		{payload: input.RegistryPayload, digest: input.RegistryDigest, reason: R4ReasonRegistryBindingMismatch, expected: projections.registry},
	}
	for _, binding := range bindings {
		computed, err := canonicalR4PayloadDigest(binding.payload)
		if err != nil {
			return R4ReasonMalformedBinding
		}
		if !bytes.Equal([]byte(binding.payload), binding.expected) || computed != binding.digest {
			return binding.reason
		}
	}
	return ""
}
