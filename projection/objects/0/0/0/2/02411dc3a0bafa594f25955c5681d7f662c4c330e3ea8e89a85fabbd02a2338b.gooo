package pressureindependence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, err
	}
	var input Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode pressure-independence input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Input{}, fmt.Errorf("decode pressure-independence input: trailing JSON value")
		}
		return Input{}, fmt.Errorf("decode pressure-independence input: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return Input{}, fmt.Errorf("decode pressure-independence input: expected object")
	}
	input.present = inputPresence{
		schema: present(fields, "schema"), fixture: present(fields, "fixture_id"),
		snapshot: present(fields, "authority_snapshot_digest"), policy: present(fields, "policy_digest"),
		registry: present(fields, "registry_digest"), oracle: present(fields, "oracle_digest"),
		toolchain: present(fields, "toolchain_options_digest"), requestedK: present(fields, "requested_K"),
		minimumIndependent: present(fields, "minimum_independent"), pressureRecords: present(fields, "pressure_records"),
		requiredIDs: present(fields, "required_pressure_ids"), guardIDs: present(fields, "guard_ids"),
		finitePaths: present(fields, "finite_path_ids"), resourceCeilings: present(fields, "resource_ceilings"),
	}
	return input, nil
}
func CanonicalInputBytes(input Input) ([]byte, error) {
	normalized := normalizeInput(input)
	return json.Marshal(normalized)
}
func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}
func CanonicalOutputDigest(output Output) string {
	view := outputDigestView{
		Schema: output.Schema, FixtureID: output.FixtureID, InputDigest: output.InputDigest,
		SelectedIDs: sortedUnique(output.SelectedIDs), UnselectedIDs: sortedUnique(output.UnselectedIDs),
		UnknownIDs: sortedUnique(output.UnknownIDs), DistinctGroupCount: output.DistinctGroupCount,
		Decision: output.Decision, Reason: output.Reason, FullSuiteRequired: output.FullSuiteRequired,
		ProofValid: output.ProofValid, CostReceipt: output.CostReceipt,
	}
	data, _ := json.Marshal(view)
	return digestBytes(data)
}
func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}
