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

func TestSequentialFillInputOriginsAndGoooConsumer4393(t *testing.T) {
	const first = `package step
namespace step4393
entity Integer id "urn:gooo:type:integer"
activity Check(Integer) -> Integer id "urn:step4393:check" computes "if input > __GOOO_BODY_HOLE_limit__ { return input + 9 }; return input - 2" assembling {
 source_fill intent "첫 번째 금액 기준" {
  hole "limit"
  candidate "step4393" { fill "limit" "31" }
  candidate "other4393" { fill "limit" "131" }
 }
 case "41" -> "50"
}
`
	const second = `package final
namespace final4393
import 앞 "chain4393/step"
activity Check(앞.Integer) -> 앞.Integer id "urn:final4393:check" computes "if input > __GOOO_BODY_HOLE_limit__ { return input + 13 }; return input - 3" assembling {
 source_fill intent "두 번째 금액 기준" {
  hole "limit"
  candidate "final4393" { fill "limit" "67" }
  candidate "other4393" { fill "limit" "167" }
 }
 case "71" -> "84"
}
bind 앞.Check.result -> Check.input
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "chain4393/final", Activity: "Check"},
		Packages: []packageruntime.PackageSpec{
			{Path: "chain4393/step", Name: "step", Sources: []packageruntime.Source{{Filename: "첫단계.gooo", Content: first}}},
			{Path: "chain4393/final", Name: "final", Imports: []string{"chain4393/step"}, Sources: []packageruntime.Source{{Filename: "끝단계.gooo", Content: second}}},
		}}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"chain4393/step:Check":79},"expected":{"chain4393/step:Check":88,"chain4393/final:Check":101}},{"inputs":{"chain4393/step:Check":23},"expected":{"chain4393/step:Check":21,"chain4393/final:Check":18}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.BodyFills) != 2 || result.Runtime.FinitePassed != 4 || len(result.Runtime.ConditionSites) != 2 {
		t.Fatal("sequential fills", result.Runtime)
	}
	if result.BodyFills[0].InputSourceSHA256 == result.BodyFills[1].InputSourceSHA256 {
		t.Fatal("staged assembly inputs collapsed")
	}
	facts := bodyexecution.ConditionFillInputFacts(result.Runtime)
	if len(facts) != 4 {
		t.Fatal("missing original file fragments", facts)
	}
	files := map[string]string{"첫단계.gooo": first, "끝단계.gooo": second}
	wants := map[string]string{"첫단계.gooo": "31", "끝단계.gooo": "67"}
	for _, fact := range facts {
		if fact.Status != "MAPPED" || fact.SourceSHA != sourceSHA256([]byte(files[fact.Filename])) {
			t.Fatal("original file origin differs", fact)
		}
		if fact.Category == "declared_candidate" && files[fact.Filename][fact.SourceStart:fact.SourceEnd] != wants[fact.Filename] {
			t.Fatal("candidate came from another file", fact)
		}
	}
	consumerSource, consumer := consumeInputOrigins4393(t, facts)
	raw, _ := json.Marshal(result)
	var saved Result
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, _, err := reconstructWorkspace(context.Background(), manifest, saved); err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("GOOO_FILL_INPUT_OUTPUT"); output != "" {
		encoded, _ := json.MarshalIndent(map[string]any{"manifest": manifest, "result": result, "facts": facts,
			"consumer_source": consumerSource, "consumer_runtime": consumer, "actual_fact_inputs": len(facts), "synthetic_fixtures": 0}, "", "  ")
		if err := os.WriteFile(filepath.Join(output, "chain4393.json"), encoded, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("chain4393: native4/4; two stages retain distinct assembly inputs and original files; Gooo interprets four actual origin fragments4/4; saved construction read; model0")
}

func consumeInputOrigins4393(t *testing.T, facts []bodyexecution.ConditionFillInputFact) (string, bodyexecution.CompositionRuntime) {
	t.Helper()
	const source = `package origins4393
namespace origins4393
entity Text id "urn:gooo:type:string"
entity Origin id "urn:origins4393:origin" fields {
 field site_id id "urn:origins4393:site" type string required one
 field activity_id id "urn:origins4393:activity" type string required one
 field candidate_id id "urn:origins4393:candidate" type string required one
 field hole_id id "urn:origins4393:hole" type string required one
 field category id "urn:origins4393:category" type string required one
 field status id "urn:origins4393:status" type string required one
 field reason id "urn:origins4393:reason" type string required one
 field package_path id "urn:origins4393:package" type string required one
 field filename id "urn:origins4393:file" type string required one
 field source_sha id "urn:origins4393:sha" type string required one
 field source_start id "urn:origins4393:start" type integer required one
 field source_end id "urn:origins4393:end" type integer required one
}
activity Describe(Origin) -> Text id "urn:origins4393:describe" computes "if input.status != \"MAPPED\" { return \"pending\" }; if input.category == \"declared_candidate\" { return input.filename + \"/candidate\" }; return input.filename + \"/template\""
`
	var rows []map[string]any
	for _, fact := range facts {
		want := fact.Filename + "/template"
		if fact.Category == "declared_candidate" {
			want = fact.Filename + "/candidate"
		}
		rows = append(rows, map[string]any{"inputs": map[string]any{"Describe": fact}, "expected": map[string]any{"Describe": want}})
	}
	raw, _ := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": rows})
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := bodyexecution.GenerateComposition(context.Background(), "출처분류.gooo", []byte(source), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := bodyexecution.ExecuteComposition(context.Background(), "출처분류.gooo", []byte(source), prior, suite, "")
	if err != nil || runtime.FinitePassed != len(facts) || runtime.ModelCalls != 0 {
		t.Fatal("Gooo origin interpretation", err, runtime)
	}
	return source, runtime
}
