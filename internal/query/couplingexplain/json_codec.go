package couplingexplain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func DecodeVerifiedEnvelope(data []byte) (VerifiedEnvelope, error) {
	var envelope VerifiedEnvelope
	if len(bytes.TrimSpace(data)) == 0 {
		return envelope, fmt.Errorf("verified envelope: empty JSON document")
	}
	if err := strictDecode(data, &envelope); err != nil {
		return VerifiedEnvelope{}, fmt.Errorf("verified envelope: %w", err)
	}
	return envelope, nil
}

func (p *PathStep) UnmarshalJSON(data []byte) error {
	var wire struct {
		FromID       string                  `json:"from_id"`
		ToID         string                  `json:"to_id"`
		Kind         semantic.InferenceKind  `json:"kind"`
		Phase        semantic.InferencePhase `json:"phase"`
		PhaseOrdinal uint64                  `json:"phase_ordinal"`
		RuleRef      string                  `json:"rule_ref,omitempty"`
		InputDigest  string                  `json:"input_digest"`
		OutputDigest string                  `json:"output_digest"`
		EvidenceRef  string                  `json:"evidence_ref,omitempty"`
	}
	if err := strictDecode(data, &wire); err != nil {
		return err
	}
	p.FromID, p.ToID, p.Kind = wire.FromID, wire.ToID, wire.Kind
	p.Phase = semantic.PhasePlacement{Phase: wire.Phase, Ordinal: wire.PhaseOrdinal}
	p.RuleRef, p.InputDigest, p.OutputDigest, p.EvidenceRef = wire.RuleRef, wire.InputDigest, wire.OutputDigest, wire.EvidenceRef
	return nil
}

func (p PathStep) MarshalJSON() ([]byte, error) {
	return json.Marshal(canonicalPathStep{
		FromID: p.FromID, ToID: p.ToID, Kind: p.Kind, Phase: p.Phase.Phase,
		PhaseOrdinal: p.Phase.Ordinal, RuleRef: p.RuleRef,
		InputDigest: p.InputDigest, OutputDigest: p.OutputDigest, EvidenceRef: p.EvidenceRef,
	})
}

func strictDecode(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, duplicate := seen[name]; duplicate {
				return fmt.Errorf("duplicate JSON key %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}
