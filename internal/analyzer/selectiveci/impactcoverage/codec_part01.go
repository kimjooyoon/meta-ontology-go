package impactcoverage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
)

type inputWire struct {
	Schema string                `json:"schema"`
	Base   *selectiveci.Snapshot `json:"base"`
	Head   *selectiveci.Snapshot `json:"head"`
}

func (input Input) CanonicalJSON() ([]byte, error) {
	if input.Schema != SchemaV1 {
		return nil, fmt.Errorf("impact coverage schema %q is invalid", input.Schema)
	}
	if input.Base == nil || input.Head == nil {
		return nil, fmt.Errorf("base and head snapshots are required")
	}
	if err := input.Base.Validate(); err != nil {
		return nil, fmt.Errorf("base snapshot: %w", err)
	}
	if err := input.Head.Validate(); err != nil {
		return nil, fmt.Errorf("head snapshot: %w", err)
	}
	return json.Marshal(inputWire{Schema: input.Schema, Base: input.Base, Head: input.Head})
}
func (input Input) Digest() string {
	data, err := input.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}
func (input Input) MarshalJSON() ([]byte, error) { return input.CanonicalJSON() }
func (input *Input) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeInput(data)
	if err != nil {
		return err
	}
	*input = decoded
	return nil
}
func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire inputWire
	if err := decoder.Decode(&wire); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	input := Input{Schema: wire.Schema, Base: wire.Base, Head: wire.Head}
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return Input{}, err
	}
	if !bytes.Equal(canonical, data) {
		return Input{}, fmt.Errorf("impact coverage input is not canonical")
	}
	return input, nil
}
