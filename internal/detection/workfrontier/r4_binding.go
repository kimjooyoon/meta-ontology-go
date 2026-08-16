package workfrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type r4Binding struct {
	payload  string
	digest   string
	reason   string
	expected []byte
}

type r4SnapshotProjection struct {
	SchemaVersion string            `json:"schema_version"`
	Roots         []string          `json:"root_obligation_ids"`
	States        []ObligationState `json:"states"`
	Paths         []RepairPath      `json:"paths"`
	IterationUses []r4IterationUse  `json:"iteration_uses"`
}

type r4IterationUse struct {
	SCCDigest      string `json:"scc_digest"`
	IterationsUsed uint64 `json:"iterations_used"`
}

type r4PolicyProjection struct {
	SchemaVersion            string                 `json:"schema_version"`
	MinimumSelectedPressures uint32                 `json:"minimum_selected_pressures"`
	Capacity                 Capacity               `json:"capacity"`
	MaxIterations            []r4MaxIterationPolicy `json:"max_iterations"`
}

type r4MaxIterationPolicy struct {
	SCCDigest     string `json:"scc_digest"`
	MaxIterations uint64 `json:"max_iterations"`
}

type r4RegistryProjection struct {
	SchemaVersion string     `json:"schema_version"`
	Pressures     []Pressure `json:"pressures"`
}

// BindR4Payloads derives the three exact canonical projection payloads and
// their SHA-256 bindings from the normalized R4 input.
func BindR4Payloads(input R4Input) (R4Input, error) {
	input = normalizeR4Input(input)
	projections, err := r4ProjectionBytes(input)
	if err != nil {
		return R4Input{}, err
	}
	input.SnapshotPayload, input.SnapshotDigest = r4PayloadBinding(projections.snapshot)
	input.PolicyPayload, input.PolicyDigest = r4PayloadBinding(projections.policy)
	input.RegistryPayload, input.RegistryDigest = r4PayloadBinding(projections.registry)
	return input, nil
}

type r4ProjectionSet struct {
	snapshot []byte
	policy   []byte
	registry []byte
}

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

func canonicalR4PayloadDigest(payload string) (string, error) {
	raw := []byte(payload)
	if len(raw) == 0 || !utf8.Valid(raw) {
		return "", fmt.Errorf("payload is not valid UTF-8")
	}
	if err := rejectR4DuplicateKeys(raw); err != nil {
		return "", fmt.Errorf("payload object: %w", err)
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("payload JSON: %w", err)
	}
	if value == nil {
		return "", fmt.Errorf("payload must be a JSON object")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", fmt.Errorf("canonical payload: %w", err)
	}
	if !bytes.Equal(compact.Bytes(), raw) {
		return "", fmt.Errorf("payload is not canonical")
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}
