package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

func computationPackageReference(raw []byte, profile packageProfile) ([]packageProfileCase, error) {
	known := false
	for _, candidate := range computationProfiles {
		known = known || profile == candidate
	}
	if !known {
		return nil, fmt.Errorf("unknown computation package profile")
	}
	var suite struct {
		Schema string
		Cases  []packageProfileCase
	}
	if err := json.Unmarshal(raw, &suite); err != nil {
		return nil, err
	}
	if suite.Schema != "gooo/body-composition-cases/v1" || len(suite.Cases) != profile.Total {
		return nil, fmt.Errorf("computation package case inventory differs")
	}
	key := profile.Package + ":" + profile.Activity
	for i, row := range suite.Cases {
		input, present := row.Inputs[key]
		expected, hasExpected := row.Expected[key]
		if !present || !hasExpected || len(row.Inputs) != 1 || len(row.Expected) != 1 {
			return nil, fmt.Errorf("computation package row %d is incomplete", i)
		}
		want, err := computationPackageValue(input, profile)
		if err != nil || !samePackageSourceValue(want, expected) {
			return nil, fmt.Errorf("computation package row %d differs from Go reference: %v", i, err)
		}
	}
	return suite.Cases, nil
}
func computationPackageValue(raw []byte, profile packageProfile) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || len(fields) != 2 ||
		len(fields["text"]) == 0 || len(fields["needle"]) == 0 || string(fields["text"]) == "null" ||
		string(fields["needle"]) == "null" {
		return nil, fmt.Errorf("computation package query requires text and needle")
	}
	var query struct {
		Text   string
		Needle json.RawMessage
	}
	if err := json.Unmarshal(raw, &query); err != nil || len(query.Text) > 1024 || !utf8.ValidString(query.Text) {
		return nil, fmt.Errorf("computation package Text differs from its contract")
	}
	if profile.Name == "single-condition-byte" {
		var needle int
		if err := json.Unmarshal(query.Needle, &needle); err != nil {
			return nil, err
		}
		offset := -1
		if needle >= 0 && needle <= 255 {
			offset = strings.IndexByte(query.Text, byte(needle))
		}
		return json.Marshal(map[string]int{"offset": offset})
	}
	if profile.Name != "string-operations" {
		return nil, fmt.Errorf("unknown computation package value profile")
	}
	var needle string
	if err := json.Unmarshal(query.Needle, &needle); err != nil || len(needle) > 1024 || !utf8.ValidString(needle) {
		return nil, fmt.Errorf("computation package needle differs from its contract")
	}
	return json.Marshal(map[string]int{"index": strings.Index(query.Text, needle), "count": strings.Count(query.Text, needle), "runes": utf8.RuneCountInString(query.Text)})
}
