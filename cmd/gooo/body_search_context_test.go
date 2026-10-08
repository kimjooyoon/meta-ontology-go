package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const searchContextSource = `package offset
namespace offset
entity Integer id "offset://integer"
entity Note id "offset://note" fields {
    field text id "offset://note/text" type string required one
}
activity Add(Integer) -> Integer computes "return input + __GOOO_BODY_HOLE_offset__" assembling {
    search hole "offset" grammar "integer-hole-residual/v1" intent "Add one." max_candidates "16"
    case "2" -> "3"
    attempts "1"
}`

func TestBodyContextExportsSourceSearchWithoutRunningCandidates(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "http://invalid.example/never")
	for _, body := range []string{"input + __GOOO_BODY_HOLE_offset__", "__GOOO_BODY_HOLE_offset__ + 1 / 0"} {
		for _, include := range []bool{false, true} {
			source := strings.Replace(searchContextSource, "input + __GOOO_BODY_HOLE_offset__", body, 1)
			args := []string{"--activity", "Add", "source.gooo"}
			if include {
				args = append(args, "--include-plan")
			}
			var out, stderr bytes.Buffer
			code := runBodyContext(args, mapSourceReader{"source.gooo": []byte(source)}, &out, &stderr)
			var result bodycodegen.SourceIRSearchContextExport
			var fields map[string]json.RawMessage
			if code != exitOK || json.Unmarshal(out.Bytes(), &result) != nil || json.Unmarshal(out.Bytes(), &fields) != nil {
				t.Fatalf("source search context unavailable: %s %s", out.String(), stderr.String())
			}
			if result.Schema != "gooo/source-search-input-export/v1" || result.ActivityID != "offset://activity/add" || len(result.OriginalSourceSHA256) != 71 || len(result.ContractSHA256) != 71 || result.ModelPredictions != 0 || result.CandidateTests != 0 || (result.ExpandedPlan != nil) != include {
				t.Fatal("source search export lost identity, plan or zero-work observations", out.String())
			}
			for _, name := range []string{"context", "candidates", "attempts", "training_case_results"} {
				if fields[name] != nil {
					t.Fatal("input export includes model features or candidate outcomes", name)
				}
			}
		}
	}
}

func TestBodyContextSourceSearchRejectsUnsupportedOptionsAndCancellation(t *testing.T) {
	reader := mapSourceReader{"source.gooo": []byte(searchContextSource), "empty.json": {}}
	for _, extra := range [][]string{{"--plan", "empty.json"}, {"--value-flow"}, {"--feature-version", decision.SplitContextIntentFeatureVersion}} {
		var out, stderr bytes.Buffer
		if runBodyContext(append([]string{"--activity", "Add", "source.gooo"}, extra...), reader, &out, &stderr) != exitFailure {
			t.Fatal("unsupported search option accepted", extra, out.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, stderr bytes.Buffer
	if runBodyContextWithContext(ctx, []string{"--activity", "Add", "source.gooo"}, reader, &out, &stderr) != exitFailure || !strings.Contains(out.String(), "canceled") {
		t.Fatal("canceled source search exported a plan", out.String())
	}
}
