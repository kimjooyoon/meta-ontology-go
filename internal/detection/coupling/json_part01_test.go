package coupling

import (
	"bytes"
	"testing"
)

func TestStrictJSONInputAndResultCodecs(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	data, err := EncodeInput(fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	result := EvaluateJSON(data, fixture.authorityContext)
	if result.Status != StatusPass || !validDigest(result.InputDigest) {
		t.Fatalf("JSON result = %#v", result)
	}
	resultData, err := EncodeResult(result)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeResult(resultData); err != nil {
		t.Fatalf("DecodeResult: %v", err)
	}
	mutations := map[string][]byte{
		"duplicate key":  []byte(`{"schema":"x","schema":"y"}`),
		"unknown field":  []byte(`{"unknown":true}`),
		"trailing value": append(append([]byte(nil), data...), []byte(`{}`)...),
		"wrong type":     []byte(`{"schema":7}`),
		"wrong schema":   bytes.Replace(data, []byte(InputSchemaV1), []byte("wrong/input/v9"), 1),
	}
	for name, mutated := range mutations {
		t.Run(name, func(t *testing.T) {
			got := EvaluateJSON(mutated, fixture.authorityContext)
			if got.Status != StatusFailClosed || got.Reasons[0].Code != ReasonMalformedBinding || !validDigest(got.InputDigest) {
				t.Fatalf("mutation result = %#v", got)
			}
		})
	}
}
func TestStrictJSONRejectsNestedDuplicateAndUnknownFields(t *testing.T) {
	fixture := newFixture(t, ChangeClaimNoDelta)
	data, err := EncodeInput(fixture.input)
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2][]byte{
		{[]byte(`"baseline":{"schema"`), []byte(`"baseline":{"unknown":true,"schema"`)},
		{[]byte(`"binding":{"source_map_id"`), []byte(`"binding":{"unknown":true,"source_map_id"`)},
	} {
		mutated := bytes.Replace(data, pair[0], pair[1], 1)
		if got := EvaluateJSON(mutated, fixture.authorityContext); got.Status != StatusFailClosed || got.Reasons[0].Code != ReasonMalformedBinding {
			t.Fatalf("nested mutation result = %#v", got)
		}
	}
}
