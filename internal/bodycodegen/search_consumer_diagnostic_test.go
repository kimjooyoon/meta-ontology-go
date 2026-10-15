package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestSearchConsumerDiagnosticPrecedesUnfilledBodyProjection(t *testing.T) {
	source := []byte(`package modes
namespace modes145
entity Integer id "urn:modes145:integer"
activity Double(Integer) -> Integer computes "return __GOOO_BODY_HOLE_seed__*2" assembling {
 search hole "seed" grammar "integer-offset-constant/v1" intent "Choose a bounded seed." max_candidates "3"
 case "7" -> "14"
 holdout_case "19" -> "38"
 attempts "2"
}
`)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "search-consumer.gooo", source, "Double")
	if err == nil || !strings.Contains(err.Error(), "search requires GenerateWithSourceIRSearch") ||
		strings.Contains(err.Error(), "undefined:") || r.Source != "" || r.GoooSource != "" {
		t.Fatal("wrong consumer projected an unfilled body or returned no usable diagnostic", err)
	}
	t.Log("SEARCH_CONSUMER_DIAGNOSTIC_OBSERVATION expected_consumer=GenerateWithSourceIRSearch code_emitted=0 model_predictions=0 candidate_scoring=0 program_executions=0")
}
