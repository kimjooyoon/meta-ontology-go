package cliobservations

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBoundedTextPublicRefusalsFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, input, reason string
		source              []byte
		native              bool
	}{
		{"ingress", strings.Repeat("x", 1025), "1024", source, false},
		{"legacy-output", strings.Repeat("x", 1024), "1024", bytes.Replace(source, []byte(` max_utf8_bytes "1030"`), nil, 1), true},
		{"above-output", strings.Repeat("x", 1024), "1030", bytes.Replace(source, []byte(`"prefix"`), []byte(`"prefix!"`), 1), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input, _ := json.Marshal(tc.input)
			suite := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
				Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{"app/bound1905:Prefix": input}}}}
			receipt, code := executeTextBound1905(t, tc.name, tc.source, "Prefix", suite, true)
			if code == exitOK || !strings.Contains(receipt.Error, tc.reason) {
				t.Fatal("source-bound refusal missing", code, receipt.Error)
			}
			if tc.name == "above-output" && !strings.Contains(receipt.Error, "gooo://examples/bounded-text/expanded/value") {
				t.Fatal("bound refusal lost its source field identity", receipt.Error)
			}
			if receipt.Result == nil {
				if tc.native {
					t.Fatal("output refusal did not retain native result")
				}
				return
			}
			r := receipt.Result.Runtime
			if r.RuntimeReplayed || r.FinitePassed != 0 || r.ModelCalls != 0 {
				t.Fatal("rejected value gained completion credit", r)
			}
			if tc.native {
				if !r.Build.Completed || len(r.Runs) != 1 || !r.Runs[0].Completed || r.Runs[0].ExitCode == nil || *r.Runs[0].ExitCode != 0 {
					t.Fatal("output boundary did not observe the original native process", r)
				}
			} else if r.Build.Started || len(r.Runs) != 0 {
				t.Fatal("invalid ingress reached native execution", r)
			}
		})
	}
}

func TestSourceDeclaredLargerTextDomainFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.Replace(source, []byte(`max_utf8_bytes "1030"`), []byte(`max_utf8_bytes "1031"`), 1)
	source = bytes.Replace(source, []byte(`"prefix"`), []byte(`"prefix!"`), 1)
	input, _ := json.Marshal(strings.Repeat("x", 1024))
	expected, _ := json.Marshal(map[string]string{"value": "prefix!" + strings.Repeat("x", 1024)})
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs:   map[string]json.RawMessage{"app/bound1905:Prefix": input},
		Expected: map[string]json.RawMessage{"app/bound1905:Prefix": expected},
	}}}
	receipt, code := executeTextBound1905(t, "declared1031", source, "Prefix", suite, false)
	assertTextBoundNative1905(t, receipt, code, 1)
}
