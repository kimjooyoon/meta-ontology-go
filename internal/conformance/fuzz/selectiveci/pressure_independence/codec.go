package pressureindependence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
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

func EncodeOutputJSON(output Output) ([]byte, error) {
	output.SelectedIDs = sortedUnique(output.SelectedIDs)
	output.UnselectedIDs = sortedUnique(output.UnselectedIDs)
	output.UnknownIDs = sortedUnique(output.UnknownIDs)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

type outputDigestView struct {
	Schema             string      `json:"schema"`
	FixtureID          string      `json:"fixture_id"`
	InputDigest        string      `json:"input_digest"`
	SelectedIDs        []string    `json:"selected_ids"`
	UnselectedIDs      []string    `json:"unselected_ids"`
	UnknownIDs         []string    `json:"unknown_ids"`
	DistinctGroupCount uint64      `json:"distinct_group_count"`
	Decision           Decision    `json:"decision"`
	Reason             Reason      `json:"reason"`
	FullSuiteRequired  bool        `json:"full_suite_required"`
	ProofValid         bool        `json:"proof_valid"`
	CostReceipt        CostReceipt `json:"cost_receipt"`
}

func normalizeInput(input Input) Input {
	input.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	input.RequiredPressureIDs = append([]string(nil), input.RequiredPressureIDs...)
	input.GuardIDs = append([]string(nil), input.GuardIDs...)
	input.FinitePathIDs = append([]string(nil), input.FinitePathIDs...)
	sort.Slice(input.PressureRecords, func(i, j int) bool {
		left, right := input.PressureRecords[i], input.PressureRecords[j]
		return pressureKey(left) < pressureKey(right)
	})
	sort.Strings(input.RequiredPressureIDs)
	sort.Strings(input.GuardIDs)
	sort.Strings(input.FinitePathIDs)
	return input
}

func pressureKey(record PressureRecord) string {
	return record.PressureID + "\x00" + record.CategoryID + "\x00" +
		record.IndependenceGroupID + "\x00" + record.ApplicabilityRuleID
}

func sortedUnique(values []string) []string {
	if values == nil {
		return []string{}
	}
	values = append([]string(nil), values...)
	if len(values) == 0 {
		return []string{}
	}
	sort.Strings(values)
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func decodeStrictJSON(data []byte, target any) error {
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
		return err
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate JSON key %q", key)
		}
		seen[key] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func present(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
