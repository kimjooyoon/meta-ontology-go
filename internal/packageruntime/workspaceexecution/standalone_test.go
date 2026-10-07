package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func standaloneWorkspace(source, entry string) packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/tool", Activity: entry},
		Packages: []packageruntime.PackageSpec{{Path: "example/tool", Name: "tool",
			Sources: []packageruntime.Source{{Filename: "tool.gooo", Content: source}}}}}
}

func TestPrepareStandalonePackageKeepsEntryWithoutSyntheticBinding(t *testing.T) {
	manifest := standaloneWorkspace(`package tool
namespace tool
entity Integer id "standalone://integer"
activity Difference(Integer, Integer) -> Integer computes "return input0 - input1"
activity Unused(Integer) -> Integer computes "return input"
`, "Difference")
	program, err := Prepare(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Activities) != 1 || program.Entry.Activity != "Difference" || strings.Contains(program.Source, "bind ") || strings.Contains(program.Source, "Unused") {
		t.Fatal("independent entry did not preserve its exact closure", program.Source)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/tool:Difference.input0":9007199254740993,"example/tool:Difference.input1":2},"expected":{"example/tool:Difference":9007199254740991}},
{"inputs":{"example/tool:Difference.input0":0,"example/tool:Difference.input1":3},"expected":{"example/tool:Difference":-3}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || !result.Runtime.RuntimeReplayed || len(result.Composition.Plan.Edges) != 0 {
		t.Fatal("standalone multi-input native execution", err, result.Runtime)
	}
	suite.Cases[0].Inputs["example/tool:Unused"] = json.RawMessage(`1`)
	if _, err := ExecuteWorkspace(context.Background(), manifest, suite, "", ""); err == nil {
		t.Fatal("input for an unreachable activity was accepted")
	}
}

func TestStandalonePackageSourceFillUsesDeclaredCandidates(t *testing.T) {
	manifest := standaloneWorkspace(`package tool
namespace tool
entity Integer id "standalone://integer"
activity Lift(Integer) -> Integer computes "let base = __GOOO_BODY_HOLE_seed__; return base + __GOOO_BODY_HOLE_step__" assembling {
    source_fill intent "Add one to the input." {
        hole "seed"
        hole "step"
        candidate "add_one" { fill "seed" "input + 0" fill "step" "1" }
        candidate "subtract" { fill "seed" "input - 0" fill "step" "-1" }
    }
    case "0" -> "1"
    case "2" -> "3"
}
`, "Lift")
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/tool:Lift":9},"expected":{"example/tool:Lift":10}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.BodyFills) != 1 || result.BodyFills[0].Generation.Report.BodyFill.SelectedCandidateID != "add_one" || result.Runtime.FinitePassed != 1 || !result.Runtime.RuntimeReplayed {
		t.Fatal("standalone source assembly did not reach native execution", result)
	}
}

func TestPrepareStandalonePackageRejectsUnresolvedEntryAndSelfBinding(t *testing.T) {
	source := "package tool\nnamespace tool\nentity Integer id \"standalone://integer\"\nactivity Main(Integer) -> Integer computes \"return input\"\n"
	if _, err := Prepare(standaloneWorkspace(source, "Missing")); err == nil {
		t.Fatal("missing entry was accepted")
	}
	manifest := standaloneWorkspace(source+"bind Main.result -> Main.input\n", "Main")
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{},"expected":{"example/tool:Main":1}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ExecuteWorkspace(context.Background(), manifest, suite, "", ""); err == nil {
		t.Fatal("self-binding was accepted")
	}
}
