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

func (p *RepairPath) UnmarshalJSON(data []byte) error {
	type wire RepairPath
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*p = RepairPath(raw)
	p.stableIDPresent = present(fields, "stable_id")
	p.obligationIDPresent = present(fields, "obligation_id")
	p.prerequisiteObligationIDsPresent = present(fields, "prerequisite_obligation_ids")
	p.readSetPresent = present(fields, "read_set")
	p.writeSetPresent = present(fields, "write_set")
	p.requiredPressureIDsPresent = present(fields, "required_pressure_ids")
	p.policyPriorityPresent = present(fields, "policy_priority")
	p.cpuCoreNSUpperBoundPresent = present(fields, "cpu_core_ns_upper_bound")
	p.fromJSON = true
	return nil
}

func (c *Capacity) UnmarshalJSON(data []byte) error {
	type wire Capacity
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*c = Capacity(raw)
	c.cpuCoreNSPresent = present(fields, "cpu_core_ns")
	return nil
}

func (in *Input) UnmarshalJSON(data []byte) error {
	type wire Input
	var raw wire
	fields, err := decodeStrictObject(data, &raw)
	if err != nil {
		return err
	}
	*in = Input(raw)
	in.fromJSON = true
	in.present = inputPresence{
		schemaVersion:            present(fields, "schema_version"),
		snapshotDigest:           present(fields, "snapshot_digest"),
		policyDigest:             present(fields, "policy_digest"),
		registryDigest:           present(fields, "registry_digest"),
		minimumSelectedPressures: present(fields, "minimum_selected_pressures"),
		capacity:                 present(fields, "capacity"),
		pressures:                present(fields, "pressures"),
		states:                   present(fields, "states"),
		paths:                    present(fields, "paths"),
	}
	return nil
}

// DecodeJSON parses one strict work-frontier input object.
func DecodeJSON(data []byte) (Input, error) {
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode work frontier JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Input{}, fmt.Errorf("decode work frontier JSON: multiple values")
		}
		return Input{}, fmt.Errorf("decode work frontier JSON: %w", err)
	}
	return input, nil
}

// Decode is the JSON entry point for the v1 contract.
func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

// EncodeJSON returns canonical, indented JSON terminated by one newline.
func EncodeJSON(input Input) ([]byte, error) {
	normalized := normalizeInput(input)
	encoded, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode work frontier JSON: %w", err)
	}
	return append(encoded, '\n'), nil
}

// EncodeResultJSON returns canonical JSON for a selection result.
func EncodeResultJSON(result Result) ([]byte, error) {
	normalized := normalizeResult(result)
	encoded, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode work frontier result: %w", err)
	}
	return append(encoded, '\n'), nil
}
