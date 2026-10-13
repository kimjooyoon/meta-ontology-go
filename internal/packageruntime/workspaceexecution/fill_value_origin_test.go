package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestCandidateThroughLocalValue5107(t *testing.T) {
	const source = `package eligibility
namespace eligibility5107
entity Integer id "urn:gooo:type:integer"
activity Calculate(Integer) -> Integer id "urn:eligibility5107:calculate" computes "let eligible = input > __GOOO_BODY_HOLE_floor__; let amount = input; if eligible { amount = input + 17 } else { amount = input - 4 }; return amount" assembling {
 source_fill intent "기준보다 크면 금액에 17을 더하고, 그렇지 않으면 4를 뺀다" {
  hole "floor"
  candidate "floor5107" { fill "floor" "53" }
  candidate "other5107" { fill "floor" "153" }
 }
 case "61" -> "78"
}
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "value5107", Activity: "Calculate"},
		Packages: []packageruntime.PackageSpec{{Path: "value5107", Name: "eligibility",
			Sources: []packageruntime.Source{{Filename: "자격.gooo", Content: source}}}}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"value5107:Calculate":69},"expected":{"value5107:Calculate":86}},{"inputs":{"value5107:Calculate":41},"expected":{"value5107:Calculate":37}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Runtime.FinitePassed != 2 || len(result.Runtime.ConditionSites) != 1 {
		t.Fatal("native observations", result.Runtime)
	}
	facts := bodyexecution.ConditionDependencyFacts(result.Runtime)
	if len(facts) != 1 {
		t.Fatal("candidate behind local assignment missing", facts)
	}
	fact := facts[0]
	if fact.CandidateID != "floor5107" || fact.HoleID != "floor" || fact.Relation != "static_dependency" || fact.AnalysisStatus != "MAPPED" || fact.SourceStatus != "MAPPED" || fact.Filename != "자격.gooo" || source[fact.SourceStart:fact.SourceEnd] != "53" {
		t.Fatal("original candidate lineage", fact)
	}
	consumerSource, consumer := consumeDependencies5107(t, facts)
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var saved Result
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		t.Fatal("strict receipt", err)
	}
	if _, _, err := reconstructWorkspace(context.Background(), manifest, saved); err != nil {
		t.Fatal("saved construction", err)
	}
	if output := os.Getenv("GOOO_VALUE_ORIGIN_OUTPUT"); output != "" {
		encoded, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "result": result, "facts": facts, "consumer_source": consumerSource, "consumer_runtime": consumer, "synthetic_fixtures": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "value5107.json"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("5107: native2/2; candidate -> local variable -> condition -> original package file; Gooo reads one actual dependency1/1; strict saved construction read; model0")
}

func consumeDependencies5107(t *testing.T, facts []bodyexecution.ConditionDependencyFact) (string, bodyexecution.CompositionRuntime) {
	t.Helper()
	const source = `package originview
namespace originview5107
entity Text id "urn:gooo:type:string"
entity Origin id "urn:originview5107:origin" fields {
 field site_id id "urn:originview5107:site" type string required one
 field activity_id id "urn:originview5107:activity" type string required one
 field candidate_id id "urn:originview5107:candidate" type string required one
 field hole_id id "urn:originview5107:hole" type string required one
 field relation id "urn:originview5107:relation" type string required one
 field analysis_status id "urn:originview5107:analysis" type string required one
 field source_status id "urn:originview5107:source" type string required one
 field filename id "urn:originview5107:file" type string required one
 field source_sha id "urn:originview5107:sha" type string required one
 field source_start id "urn:originview5107:start" type integer required one
 field source_end id "urn:originview5107:end" type integer required one
}
activity Explain(Origin) -> Text id "urn:originview5107:explain" computes "if input.analysis_status != \"MAPPED\" { return \"부분 추적\" }; if input.relation == \"static_dependency\" { return input.filename + \"에서 변수를 거쳐 연결됨\" }; return input.filename"
`
	var rows []map[string]any
	for _, fact := range facts {
		rows = append(rows, map[string]any{"inputs": map[string]any{"Explain": fact}, "expected": map[string]any{"Explain": fact.Filename + "에서 변수를 거쳐 연결됨"}})
	}
	raw, _ := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodyexecution.GenerateComposition(context.Background(), "기원설명.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := bodyexecution.ExecuteComposition(context.Background(), "기원설명.gooo", []byte(source), prior, suite, "")
	if err != nil || runtime.FinitePassed != len(facts) || runtime.ModelCalls != 0 {
		t.Fatal("Gooo dependency interpretation", err, runtime)
	}
	return source, runtime
}
