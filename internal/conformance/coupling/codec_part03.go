package coupling

import (
	"encoding/json"
)

func EncodeInputJSON(input Input) ([]byte, error) {
	raw := inputToWire(input, false)
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// CanonicalInputBytes excludes fixture labels. The result is the exact byte
// sequence used for InputDigest, so fixture expectations cannot affect it.
func CanonicalInputBytes(input Input) ([]byte, error) {
	raw := inputToWire(input, true)
	return json.Marshal(raw)
}
func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}

type outputDigestView struct {
	Schema               string              `json:"schema"`
	InputDigest          string              `json:"input_digest"`
	Decision             Decision            `json:"decision"`
	Reason               Reason              `json:"reason"`
	AcceptedSurfaces     []string            `json:"accepted_surfaces"`
	ChangedSurfaces      []string            `json:"changed_surfaces"`
	ReceiptSurfaces      []string            `json:"receipt_surfaces"`
	SemanticBeforeDigest string              `json:"semantic_before_digest"`
	SemanticAfterDigest  string              `json:"semantic_after_digest"`
	SemanticDeltaDigest  string              `json:"semantic_delta_digest,omitempty"`
	PathClosureDigest    string              `json:"path_closure_digest,omitempty"`
	ObservationCounts    ObservationCounts   `json:"observation_counts"`
	Resources            ResourceObservation `json:"resources"`
}

func CanonicalOutputDigest(output Output) string {
	view := outputDigestView{
		Schema: output.Schema, InputDigest: output.InputDigest, Decision: output.Decision, Reason: output.Reason,
		AcceptedSurfaces: sortedUnique(output.AcceptedSurfaces),
		ChangedSurfaces:  sortedUnique(output.ChangedSurfaces), ReceiptSurfaces: sortedUnique(output.ReceiptSurfaces),
		SemanticBeforeDigest: output.SemanticBeforeDigest, SemanticAfterDigest: output.SemanticAfterDigest,
		SemanticDeltaDigest: output.SemanticDeltaDigest, PathClosureDigest: output.PathClosureDigest,
		ObservationCounts: output.ObservationCounts, Resources: output.Resources,
	}
	data, _ := json.Marshal(view)
	return digestBytes(data)
}
func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}
func EncodeOutputJSON(output Output) ([]byte, error) {
	output.ChangedSurfaces = sortedUnique(output.ChangedSurfaces)
	output.ReceiptSurfaces = sortedUnique(output.ReceiptSurfaces)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
