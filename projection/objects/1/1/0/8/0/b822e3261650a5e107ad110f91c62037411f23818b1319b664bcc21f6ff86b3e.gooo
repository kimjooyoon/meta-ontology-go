package workfrontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func decodeStrictObject(data []byte, target any) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("work frontier object: %w", err)
	}
	if fields == nil {
		return nil, fmt.Errorf("work frontier object: expected an object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, fmt.Errorf("work frontier object: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("work frontier object: multiple values")
		}
		return nil, fmt.Errorf("work frontier object: %w", err)
	}
	return fields, nil
}
func present(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func (p *Pressure) UnmarshalJSON(data []byte) error {
	type wire Pressure
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*p = Pressure(raw)
	p.stableIDPresent = present(fields, "stable_id")
	p.fromJSON = true
	return nil
}
func (s *ObligationState) UnmarshalJSON(data []byte) error {
	type wire ObligationState
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*s = ObligationState(raw)
	s.obligationIDPresent = present(fields, "obligation_id")
	s.statusPresent = present(fields, "status")
	s.fromJSON = true
	return nil
}
