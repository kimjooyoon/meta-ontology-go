package workfrontier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

type r4DecodeFailureKind uint8

const (
	r4DecodeMissingRequired r4DecodeFailureKind = iota + 1
	r4DecodeMalformed
)

type r4DecodeError struct {
	kind r4DecodeFailureKind
	err  error
}

func (e *r4DecodeError) Error() string {
	return fmt.Sprintf("decode r4 work frontier: %v", e.err)
}

func (e *r4DecodeError) Unwrap() error { return e.err }

// DecodeR4JSON accepts exactly one R4 envelope. Unknown fields, duplicate
// fields, missing required fields, and legacy schema versions are rejected.
func DecodeR4JSON(data []byte) (R4Input, error) {
	return decodeR4JSON(data)
}

func decodeR4JSON(data []byte) (R4Input, error) {
	if err := rejectR4DuplicateKeys(data); err != nil {
		return R4Input{}, &r4DecodeError{kind: r4DecodeMalformed, err: err}
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		if err == nil {
			err = fmt.Errorf("expected an object")
		}
		return R4Input{}, &r4DecodeError{kind: r4DecodeMalformed, err: err}
	}
	for _, field := range []string{
		"schema_version", "snapshot_digest", "snapshot_payload", "policy_digest", "policy_payload",
		"registry_digest", "registry_payload", "minimum_selected_pressures", "capacity", "pressures",
		"states", "paths", "root_obligation_ids", "rules",
	} {
		if !r4FieldPresent(fields, field) {
			return R4Input{}, &r4DecodeError{
				kind: r4DecodeMissingRequired,
				err:  fmt.Errorf("required field %q is missing", field),
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input R4Input
	if err := decoder.Decode(&input); err != nil {
		return R4Input{}, &r4DecodeError{kind: r4DecodeMalformed, err: err}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return R4Input{}, &r4DecodeError{kind: r4DecodeMalformed, err: err}
	}
	if input.SchemaVersion != R4SchemaVersion {
		return R4Input{}, &r4DecodeError{
			kind: r4DecodeMalformed,
			err:  fmt.Errorf("schema_version must be %q", R4SchemaVersion),
		}
	}
	return normalizeR4Input(input), nil
}

// EvaluateR4JSON maps only missing required fields to UNKNOWN. Every other
// strict-envelope decode failure is FAIL_CLOSED.
func EvaluateR4JSON(data []byte) R4Result {
	input, err := decodeR4JSON(data)
	if err != nil {
		var decodeErr *r4DecodeError
		if errors.As(err, &decodeErr) && decodeErr.kind == r4DecodeMissingRequired {
			return r4RequiredInputResult()
		}
		return r4FailClosed(r4Graph{}, R4ReasonMalformedBinding)
	}
	return EvaluateR4(input)
}

// EncodeR4JSON emits one normalized, versioned input envelope.
func EncodeR4JSON(input R4Input) ([]byte, error) {
	input = normalizeR4Input(input)
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode r4 work frontier: %w", err)
	}
	return append(data, '\n'), nil
}

// EncodeR4ResultJSON emits the deterministic result without adding any
// authorization or proof fields.
func EncodeR4ResultJSON(result R4Result) ([]byte, error) {
	result = normalizeR4Result(result)
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode r4 work frontier result: %w", err)
	}
	return append(data, '\n'), nil
}

func normalizeR4Input(input R4Input) R4Input {
	input.Pressures = append([]Pressure(nil), input.Pressures...)
	input.States = append([]ObligationState(nil), input.States...)
	input.Paths = append([]RepairPath(nil), input.Paths...)
	input.RootObligationIDs = sortedCopy(input.RootObligationIDs)
	input.Rules = normalizeR4Rules(input.Rules)
	for index := range input.Paths {
		input.Paths[index].PrerequisiteObligationIDs = sortedCopy(input.Paths[index].PrerequisiteObligationIDs)
		input.Paths[index].ReadSet = sortedCopy(input.Paths[index].ReadSet)
		input.Paths[index].WriteSet = sortedCopy(input.Paths[index].WriteSet)
		input.Paths[index].RequiredPressureIDs = sortedCopy(input.Paths[index].RequiredPressureIDs)
	}
	sort.Slice(input.Pressures, func(i, j int) bool { return input.Pressures[i].StableID < input.Pressures[j].StableID })
	sort.Slice(input.States, func(i, j int) bool {
		if input.States[i].ObligationID != input.States[j].ObligationID {
			return input.States[i].ObligationID < input.States[j].ObligationID
		}
		return input.States[i].Status < input.States[j].Status
	})
	sort.Slice(input.Paths, func(i, j int) bool {
		return r4PathKey(input.Paths[i]) < r4PathKey(input.Paths[j])
	})
	return input
}

func normalizeR4Result(result R4Result) R4Result {
	result.Selected = uniqueInOrder(result.Selected)
	result.SelectedIDs = uniqueInOrder(result.SelectedIDs)
	result.WorkIDs = uniqueInOrder(result.WorkIDs)
	result.Unknown = sortedUnique(result.Unknown)
	result.Blocked = sortedUnique(result.Blocked)
	result.Shortfall = sortedUnique(result.Shortfall)
	return result
}

func r4PathKey(path RepairPath) string {
	return path.StableID + "\x00" + path.ObligationID + "\x00" +
		joinR4(path.PrerequisiteObligationIDs) + "\x00" + joinR4(path.ReadSet) + "\x00" +
		joinR4(path.WriteSet) + "\x00" + joinR4(path.RequiredPressureIDs) + "\x00" +
		path.WorkID + "\x00" + fmt.Sprint(path.PolicyPriority) + "\x00" + fmt.Sprint(path.CPUCoreNSUpperBound)
}

func joinR4(values []string) string {
	result := ""
	for index, value := range values {
		if index != 0 {
			result += "\x00"
		}
		result += value
	}
	return result
}

func r4FieldPresent(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func rejectR4DuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanR4JSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanR4JSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
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
				return fmt.Errorf("duplicate field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanR4JSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanR4JSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}
