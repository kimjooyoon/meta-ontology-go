package workfrontier

import (
	"encoding/json"
	"testing"
)

func TestR4JSONDecodeRouting(t *testing.T) {
	valid, err := EncodeR4JSON(r4FixtureInput(t, "acyclic"))
	if err != nil {
		t.Fatal(err)
	}
	malformedPayload, err := json.Marshal(`{"duplicate":1,"duplicate":2}`)
	if err != nil {
		t.Fatal(err)
	}
	multiple := append(append([]byte(nil), valid...), valid...)
	cases := []struct {
		name   string
		data   []byte
		status string
		reason string
	}{
		{"missing_field", r4JSONWithout(t, valid, "root_obligation_ids"), R4StatusUnknown, R4ReasonRequiredInputMissing},
		{"null_required_binding", r4JSONWith(t, valid, "snapshot_payload", []byte("null")), R4StatusUnknown, R4ReasonRequiredInputMissing},
		{"empty_required_binding", r4JSONWith(t, valid, "snapshot_payload", []byte(`""`)), R4StatusUnknown, R4ReasonRequiredInputMissing},
		{"duplicate_key", []byte(`{"schema_version":"gooo/work-frontier-r4/v1","schema_version":"gooo/work-frontier-r4/v1"}`), R4StatusFailClosed, R4ReasonMalformedBinding},
		{"unknown_field", r4JSONWith(t, valid, "unexpected", []byte("true")), R4StatusFailClosed, R4ReasonMalformedBinding},
		{"invalid_json", []byte(`{"schema_version":`), R4StatusFailClosed, R4ReasonMalformedBinding},
		{"wrong_field_type", r4JSONWith(t, valid, "capacity", []byte(`"wrong"`)), R4StatusFailClosed, R4ReasonMalformedBinding},
		{"multiple_values", multiple, R4StatusFailClosed, R4ReasonMalformedBinding},
		{"wrong_schema", r4JSONWith(t, valid, "schema_version", []byte(`"gooo/work-frontier-r4/v0"`)), R4StatusFailClosed, R4ReasonMalformedBinding},
		{"malformed_payload", r4JSONWith(t, valid, "registry_payload", malformedPayload), R4StatusFailClosed, R4ReasonMalformedBinding},
	}
	for _, fixture := range cases {
		t.Run(fixture.name, func(t *testing.T) {
			got := EvaluateR4JSON(fixture.data)
			if got.Status != fixture.status || got.Reason != fixture.reason || len(got.SelectedIDs) != 0 {
				t.Fatalf("result = %#v, want %s/%s with empty selection", got, fixture.status, fixture.reason)
			}
		})
	}
}

func r4JSONWith(t *testing.T, data []byte, field string, value []byte) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	object[field] = json.RawMessage(value)
	result, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func r4JSONWithout(t *testing.T, data []byte, field string) []byte {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, field)
	result, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
