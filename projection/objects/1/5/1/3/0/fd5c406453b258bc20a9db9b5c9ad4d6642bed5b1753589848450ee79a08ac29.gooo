package adapter

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOracleCorpusDeclaresExactStableCodes(t *testing.T) {
	data, err := os.ReadFile("testdata/no-write/oracle-corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name string `json:"name"`
		Code string `json:"oracle_code"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{
		OracleNW001: true, OracleNW002: true, OracleNW003: true, OracleNW004: true,
		OracleNW005: true, OracleNW006: true, OracleFAIL001: true, OracleFAIL002: true,
		OraclePASS001: true, OracleID001: true,
	}
	seen := make(map[string]bool, len(cases))
	for _, testCase := range cases {
		if testCase.Name == "" || !expected[testCase.Code] {
			t.Fatalf("invalid oracle fixture case: %+v", testCase)
		}
		seen[testCase.Code] = true
	}
	for code := range expected {
		if !seen[code] {
			t.Fatalf("oracle fixture corpus omitted %s", code)
		}
	}
}
func TestOracleNW003RejectsUntrustedOrMalformedObserverTrace(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	for _, test := range []struct {
		name string
		edit func(*NoWriteObservation)
	}{
		{name: "ORACLE-NW-003-untrusted", edit: func(observation *NoWriteObservation) { *observation = NoWriteObservation{} }},
		{name: "ORACLE-NW-003-malformed-digest", edit: func(observation *NoWriteObservation) { observation.Before.Source.ByteDigest = "sha256:bad" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			observer := newStableObserver(t, request)
			observation, err := observer.Finish()
			if err != nil {
				t.Fatal(err)
			}
			test.edit(&observation)
			evaluation := EvaluateObserved(request, response, &observation)
			assertOracleFailure(t, evaluation, OracleNW003)
		})
	}
}
