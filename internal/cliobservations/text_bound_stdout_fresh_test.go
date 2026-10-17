package cliobservations

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestSourceTextBoundCannotWidenNativeStdoutFresh1905(t *testing.T) {
	source := `package wire_bound
namespace wire_bound
entity Text id "urn:gooo:type:string"
entity Pack id "bound1905://wire/pack" fields {
 field values id "bound1905://wire/values" type string required many max_utf8_bytes "1030"
}
activity Repeat(Text) -> Pack id "bound1905://wire/repeat" computes ` + "`" +
		`return Pack{values: []string{` + strings.Repeat("input,", 255) + "input}}`\n"
	input, _ := json.Marshal(strings.Repeat("\x00", 1024))
	suite := bodyexecution.CompositionCases{Schema: bodyexecution.CompositionInputsSchema,
		Cases: []bodyexecution.CompositionCase{{Inputs: map[string]json.RawMessage{"app/bound1905:Repeat": input}}}}
	receipt, code := executeTextBound1905(t, "stdout-budget", []byte(source), "Repeat", suite, true)
	if code == exitOK || receipt.Result == nil {
		t.Fatal("valid element domains widened the independent native wire budget", code, receipt.Error)
	}
	runtime := receipt.Result.Runtime
	if runtime.RuntimeReplayed || runtime.FinitePassed != 0 || runtime.ModelCalls != 0 ||
		!runtime.Build.Completed || len(runtime.Runs) != 1 || !runtime.Runs[0].OutputTruncated {
		t.Fatal("native overflow did not retain its original process refusal", runtime)
	}
}
