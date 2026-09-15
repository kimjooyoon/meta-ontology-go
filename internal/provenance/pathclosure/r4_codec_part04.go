package pathclosure

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func decodeStrictR4(data []byte, target any) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("strict JSON: top-level object is required")
	}
	check := json.NewDecoder(bytes.NewReader(data))
	if err := walkR4JSON(check); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	if err := requireR4EOF(check); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("strict JSON: %w", err)
	}
	return requireR4EOF(decoder)
}

// EncodeR4Input emits strict canonical JSON with stable top-level ordering.
func EncodeR4Input(value R4Input) ([]byte, error) { return json.Marshal(wireR4Input(value)) }

// DecodeR4Input rejects duplicate keys, unknown fields, trailing values, and
// non-canonical whitespace/order. Semantic validity is checked by EvaluateR4.
func DecodeR4Input(data []byte) (R4Input, error) {
	var wire r4WireInput
	if err := decodeStrictR4(data, &wire); err != nil {
		return R4Input{}, err
	}
	value := r4InputFromWire(wire)
	canonical, err := EncodeR4Input(value)
	if err != nil {
		return R4Input{}, err
	}
	if !bytes.Equal(bytes.TrimSpace(data), canonical) {
		return R4Input{}, fmt.Errorf("strict JSON: non-canonical encoding")
	}
	return value, nil
}
func (value R4Input) MarshalJSON() ([]byte, error) { return EncodeR4Input(value) }
func (value *R4Input) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeR4Input(data)
	if err != nil {
		return err
	}
	*value = decoded
	return nil
}
