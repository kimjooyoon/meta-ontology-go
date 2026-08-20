package workfrontier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

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
