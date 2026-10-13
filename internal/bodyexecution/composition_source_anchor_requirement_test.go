package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNativeSourceAnchorRequirement1447(t *testing.T) {
	source := []byte(`package anchor1447
namespace anchor1447
entity Integer id "anchor1447://integer"
activity Decide(Integer) -> Integer id "anchor1447://decide" computes "if input < 41 { return input * 2 }; return input - 6"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Decide":13},"expected":{"Decide":26}},{"inputs":{"Decide":52},"expected":{"Decide":46}}]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "anchor1447.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, "anchor1447.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || !run.RuntimeReplayed || len(run.ConditionSites) != 1 {
		t.Fatal("new native requirement unavailable", err)
	}
	saveConditionObservation(t, "anchor1447", source, run)
	raw, _ := json.Marshal(run.ConditionSites)
	var sites []struct {
		Source *struct {
			Status       string `json:"status"`
			Expression   string `json:"expression"`
			SourceSHA256 string `json:"source_sha256"`
		} `json:"source"`
	}
	if err := json.Unmarshal(raw, &sites); err != nil {
		t.Fatal(err)
	}
	if len(sites) != 1 || sites[0].Source == nil || sites[0].Source.Status != "MAPPED" || sites[0].Source.Expression != "input < 41" || sites[0].Source.SourceSHA256 != prior.SelectedSourceSHA256 {
		t.Fatalf("GOOO_SOURCE_ANCHOR_MISSING: native2/2 sites=%s", raw)
	}
}
