package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestTypedPathModeDiagnosticNamesTheSourceOwnedConsumer(t *testing.T) {
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range []struct{ source, consumer string }{
		{`package modes
namespace modes131
entity Integer id "urn:modes131:integer"
activity Fill(Integer) -> Integer computes "return __GOOO_BODY_HOLE_base__+7" assembling {
 source_fill intent "Fill the declared base." {
  hole "base"
  candidate "input" { fill "base" "input" }
  candidate "zero" { fill "base" "0" }
 }
 case "4" -> "11"
}
`, "GenerateWithSourceIRBodyFill"},
		{`package modes
namespace modes132
entity Integer id "urn:modes132:integer"
activity Fill(Integer) -> Integer computes "return __GOOO_BODY_HOLE_base__+9" assembling {
 search hole "base" grammar "integer-offset-constant/v1" intent "Search the declared base." max_candidates "4"
 case "6" -> "15"
 attempts "2"
}
`, "GenerateWithSourceIRSearch"},
	} {
		result, err := g.GenerateSourceAssembly(context.Background(), "mode.gooo", []byte(profile.source), "Fill")
		if err == nil || !strings.Contains(err.Error(), "requires "+profile.consumer) || result.Source != "" || result.GoooSource != "" {
			t.Fatal("unsupported path mode reached ordinary body compilation", profile.consumer, err)
		}
	}
	t.Log("SOURCE_MODE_DISPATCH_OBSERVATION profiles=2 expected_consumer_diagnostics=2 code_emitted=0 model_predictions=0 candidate_scoring=0 program_executions=0")
}
