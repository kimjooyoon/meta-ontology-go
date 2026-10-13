package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCollectionCandidateAssembly7523(t *testing.T) {
	body := `let 합계 = 0; for 값 in input.amounts[0:] { if __GOOO_BODY_HOLE_gate__ { 합계 = 합계 + 값 } }; return Summary{total: 합계, amounts: input.amounts}`
	source := []byte(`package fill7523
namespace fill7523
entity Basket id "urn:fill7523:basket" fields {
 field amounts id "urn:fill7523:amounts" type integer required many
}
entity Summary id "urn:fill7523:summary" fields {
 field total id "urn:fill7523:total" type integer required one
 field amounts id "urn:fill7523:originals" type integer required many
}
activity Compute(Basket) -> Summary id "urn:fill7523:compute" computes ` + strconv.Quote(body) + ` assembling {
 source_fill intent "33 이상인 항목만 더하고 원래 순서를 보존한다" {
  hole "gate"
  candidate "negative7523" { fill "gate" "값 < 0" }
  candidate "threshold7523" { fill "gate" "값 >= 33" }
 }
 value_case "[{\"amounts\":[37,-4,8]}]" -> "{\"amounts\":[37,-4,8],\"total\":37}"
 value_case "[{\"amounts\":[]}]" -> "{\"amounts\":[],\"total\":0}"
 value_case "[{\"amounts\":[-7,34,33]}]" -> "{\"amounts\":[-7,34,33],\"total\":67}"
 holdout_value_case "[{\"amounts\":[9007199254741013,32]}]" -> "{\"amounts\":[9007199254741013,32],\"total\":9007199254741013}"
}
`)
	spec, err := SourceAssembly(context.Background(), "조립7523.gooo", source, "Compute")
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithSourceIRBodyFill(context.Background(), "조립7523.gooo", source, "Compute", spec, "", "", IRBodyFillOptions{})
	if err != nil {
		t.Fatal("finite collection assembly", err)
	}
	r := result.Report.BodyFill
	if r == nil || r.SelectedCandidateID != "threshold7523" || r.TestCasesPassed != 3 || r.TestCasesTotal != 3 ||
		r.HoldoutCasesPassed != 1 || r.HoldoutCasesTotal != 1 || r.LocalModelPredictions != 0 || r.ExternalProviderCalls != 0 {
		t.Fatal("selection and separate holdout results", r)
	}
	if len(r.CandidateScores) != 2 || len(result.Report.BodyConditions) != 1 {
		t.Fatal("finite candidates or condition source missing")
	}
	bound, err := BindBodyConditionSources([]byte(result.GoooSource), result.Report.BodyConditions)
	if err != nil || len(bound) != 1 || bound[0].Status != "MAPPED" {
		t.Fatal("selected condition origin", err, bound)
	}
	link := r.SourceOrigin.ConditionOrigin(result.Report.BodyConditions[0], FillOriginRange{bound[0].DecodedStart, bound[0].DecodedEnd})
	found := false
	if link != nil {
		for _, fragment := range link.Fragments {
			if fragment.HoleID == "gate" && fragment.ExpressionSource != nil {
				found = string(source[fragment.ExpressionSource.Start:fragment.ExpressionSource.End]) == "값 >= 33"
			}
		}
	}
	if !found || !strings.Contains(result.GoooSource, "for") {
		t.Fatal("collection condition must cite selected source expression", link)
	}
	if directory := os.Getenv("GOOO_COLLECTION_OBSERVATION"); directory != "" {
		raw, err := json.MarshalIndent(map[string]any{"source": string(source), "result": result, "condition_source": bound,
			"condition_construction": link, "historical_source_executions": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "fill7523.json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("7523: two declared candidates, selected finite cases3/3, disjoint holdout1/1; ordered arrays compared by values; candidate source linked; model0")
}
