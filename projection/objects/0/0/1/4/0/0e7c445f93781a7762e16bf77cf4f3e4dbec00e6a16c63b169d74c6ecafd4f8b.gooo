package pressurecoverage

import (
	"encoding/json"
	"testing"
)

func TestCanonicalInputRejectsMalformedAndAmbiguousRecords(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Input)
	}{
		{"internal whitespace", func(input *Input) { input.RequiredPressureIDs[0] = "pressure z" }},
		{"control character", func(input *Input) { input.PressureRecords[0].CategoryID = "category\x00" }},
		{"duplicate record", func(input *Input) {
			input.PressureRecords = append(input.PressureRecords, input.PressureRecords[0])
		}},
		{"conflicting record", func(input *Input) {
			input.PressureRecords = append(input.PressureRecords, PressureRecord{
				"pressure-a", "category-a", "group-new", "rule-1",
			})
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixture()
			test.edit(&input)
			if _, err := CanonicalInputBytes(input); err == nil {
				t.Fatal("accepted malformed or ambiguous input")
			}
		})
	}
}
func TestStrictJSONBoundary(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"unknown top-level field", schemaJSON(`,"display_name":"shown"`)},
		{"unknown nested field", schemaJSON(`,"pressure_records":[{"pressure_id":"p","display_name":"shown"}]`)},
		{"duplicate top-level key", schemaJSON(`,"schema":"` + SchemaVersion + `"`)},
		{"duplicate nested key", schemaJSON(`,"pressure_records":[{"pressure_id":"p","pressure_id":"q"}]`)},
		{"trailing JSON", schemaJSON("") + " " + schemaJSON("")},
		{"invalid schema", `{"schema":"wrong"}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodeInput([]byte(test.data)); err == nil {
				t.Fatal("DecodeInput accepted invalid JSON")
			}
			var input Input
			if err := json.Unmarshal([]byte(test.data), &input); err == nil {
				t.Fatal("json.Unmarshal accepted invalid JSON")
			}
		})
	}
}
