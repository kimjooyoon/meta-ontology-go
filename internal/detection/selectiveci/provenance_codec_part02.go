package selectiveci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"io"
)

func (path ProvenancePath) MarshalJSON() ([]byte, error) {
	encoded, err := pathWireFromSemantic(path.Path)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		CommandID   string          `json:"command_id"`
		Path        pathWire        `json:"path"`
		Requirement PathRequirement `json:"requirement"`
	}{path.CommandID, encoded, path.Requirement})
}
func (path *ProvenancePath) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	var raw struct {
		CommandID   string          `json:"command_id"`
		Path        pathWire        `json:"path"`
		Requirement PathRequirement `json:"requirement"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("provenance path has trailing data")
	}
	decoded, err := pathWireToSemantic(raw.Path)
	if err != nil {
		return err
	}
	*path = ProvenancePath{CommandID: raw.CommandID, Path: decoded, Requirement: raw.Requirement}
	return nil
}
func pathWireFromSemantic(path semantic.InferencePathV1) (pathWire, error) {
	result := pathWire{Version: path.Version}
	for _, edge := range path.Edges {
		converted, err := edgeWireFromSemantic(edge)
		if err != nil {
			return pathWire{}, err
		}
		result.Edges = append(result.Edges, converted)
	}
	for _, claim := range path.Claims {
		converted, err := claimWireFromSemantic(claim)
		if err != nil {
			return pathWire{}, err
		}
		result.Claims = append(result.Claims, converted)
	}
	for _, evidence := range path.Evidence {
		converted, err := evidenceWireFromSemantic(evidence)
		if err != nil {
			return pathWire{}, err
		}
		result.Evidence = append(result.Evidence, converted)
	}
	return result, nil
}
