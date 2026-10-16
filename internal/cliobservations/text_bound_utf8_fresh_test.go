package cliobservations

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBoundedTextRefusesNativeUTF8ReplacementFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = bytes.Replace(source, []byte(`"prefix" + input`), []byte(`input[1:]`), 1)
	suite := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{"app/bound1905:Prefix": json.RawMessage(`"한"`)}}}}
	receipt, code := executeTextBound1905(t, "invalid-native-utf8", source, "Prefix", suite, true)
	if code == exitOK || receipt.Result == nil {
		t.Fatal("JSON replaced a malformed native Text value", code, receipt.Error)
	}
	runtime := receipt.Result.Runtime
	if runtime.RuntimeReplayed || runtime.FinitePassed != 0 || runtime.ModelCalls != 0 ||
		!runtime.Build.Completed || len(runtime.Runs) != 1 || runtime.Runs[0].ExitCode == nil || *runtime.Runs[0].ExitCode != 4 {
		t.Fatal("native Text encoding refusal lost its original process", runtime)
	}
}
