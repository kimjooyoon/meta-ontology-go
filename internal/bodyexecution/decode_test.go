package bodyexecution

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRuntimeJSONRejectsAmbiguousOrMissingExpectations(t *testing.T) {
	for _, raw := range []string{
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0}]}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0,"expected":null}]}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0,"expected":1,"expected":2}]}`,
		`{"schema":"gooo/body-runtime-cases/v1","Cases":[{"input":0,"expected":1}]}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":null}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0,"expected":1}]} {}`,
		`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":0,"expected":1}],"extra":0}`,
		strings.Repeat("[", 66) + "0" + strings.Repeat("]", 66),
		string([]byte{0xff}),
	} {
		if _, err := DecodeCases([]byte(raw)); err == nil {
			t.Fatalf("accepted ambiguous suite: %s", raw)
		}
	}
	cases, err := DecodeCases([]byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":-9223372036854775808,"expected":9223372036854775807}]}`))
	if err != nil || cases[0].Input != -9223372036854775808 || cases[0].Expected != 9223372036854775807 {
		t.Fatal("lost exact int64 endpoints", err)
	}
}

func TestGenerationRejectsDuplicateAliasTrailingAndOversize(t *testing.T) {
	_, _, prior, _ := fixture(t)
	raw, _ := json.Marshal(prior)
	for _, malformed := range [][]byte{
		bytes.Replace(raw, []byte(`"report":`), []byte(`"report":{},"report":`), 1),
		bytes.Replace(raw, []byte(`"source":`), []byte(`"Source":`), 1),
		append(bytes.Clone(raw), []byte(` {}`)...),
		bytes.Repeat([]byte(" "), (2<<20)+1),
	} {
		if _, _, err := DecodeGeneration(malformed); err == nil {
			t.Fatal("accepted malformed generation")
		}
	}
}
