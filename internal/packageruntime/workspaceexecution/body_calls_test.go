package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func pureCallWorkspace() packageruntime.Manifest {
	return packageruntime.Manifest{Schema: packageruntime.ManifestSchema,
		Entry: packageruntime.EntrySpec{PackagePath: "example/app", Activity: "Main"},
		Packages: []packageruntime.PackageSpec{
			{Path: "example/core", Name: "core", Sources: []packageruntime.Source{
				{Filename: "scale.gooo", Content: "package core\nnamespace core\nentity Integer id \"calls://integer\"\nactivity Scale(Integer) -> Integer computes `return Triple(input) - 1`\n"},
				{Filename: "triple.gooo", Content: "package core\nnamespace core\nactivity Triple(Integer) -> Integer computes `return input * 3`\nactivity Unused(Integer) -> Integer computes `return input`\n"}}},
			{Path: "example/app", Name: "app", Imports: []string{"example/core"}, Sources: []packageruntime.Source{
				{Filename: "app.gooo", Content: "package app\nnamespace app\nimport core \"example/core\"\n" +
					"activity Main(Integer) -> Integer computes `let note = \"core.Scale(input)\"; if note == \"\" { return input }; return core.Scale(Double(input))`\n" +
					"activity Double(Integer) -> Integer computes `return input * 2`\n"}}},
		}}
}

func pureCallCases(t *testing.T) bodyexecution.CompositionCases {
	t.Helper()
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"example/app:Main":4},"expected":{"example/app:Main":23}},
{"inputs":{"example/app:Main":-2},"expected":{"example/app:Main":-13}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func TestWorkspacePureCallsNamesNativeAndSavedReplay(t *testing.T) {
	manifest, suite := pureCallWorkspace(), pureCallCases(t)
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	p := result.Program
	if len(p.Activities) != 1 || p.PureCalls == nil || len(p.PureCalls.Activities) != 4 || len(p.PureCalls.Sites) != 3 ||
		!strings.Contains(p.Source, `core.Scale(input)`) || strings.Contains(p.Source, "ActivityUnused") {
		t.Fatal("callee closure or source literal differs", p)
	}
	if result.Runtime.FinitePassed != 2 || len(result.Composition.Plan.Activities) != 1 || !result.Runtime.RuntimeReplayed {
		t.Fatal("helper became a separately supplied activity", result.Runtime)
	}
	for _, a := range p.PureCalls.Activities {
		if a.ActivityID == "" || a.LoweredID == "" || a.SourceBodySHA256 == "" || a.Source == "" {
			t.Fatal("missing source identity", a)
		}
	}
	raw, _ := json.Marshal(result)
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayWorkspace(context.Background(), manifest, saved, suite, "")
	if err != nil || replayed.Runtime.FinitePassed != 2 || replayed.Replay.ModelCalls != 0 {
		t.Fatal("package call replay", err, replayed.Runtime)
	}
	manifest.Packages[0].Sources[1].Content = strings.Replace(manifest.Packages[0].Sources[1].Content, "input * 3", "input * 4", 1)
	if _, err := ReplayWorkspace(context.Background(), manifest, saved, suite, ""); err == nil {
		t.Fatal("changed imported helper body reused old receipt")
	}
}

func TestWorkspacePureCallsRejectUnresolvedAndShadowedNames(t *testing.T) {
	for _, body := range []string{"return missing.Scale(input)", "return core.Missing(input)", "return Triple(input)",
		"let core = input; return core.Scale(input)", "let Double = input; return Double(input)", "return input.Scale(input)", "return int64(input)"} {
		t.Run(body, func(t *testing.T) {
			manifest := pureCallWorkspace()
			manifest.Packages[1].Sources[0].Content = strings.Replace(manifest.Packages[1].Sources[0].Content,
				"let note = \"core.Scale(input)\"; if note == \"\" { return input }; return core.Scale(Double(input))", body, 1)
			if _, err := Prepare(manifest); err == nil {
				t.Fatal("invalid call accepted", body)
			}
		})
	}
}

func TestWorkspacePureCallsKeepExplicitBindProducers(t *testing.T) {
	manifest, suite := pureCallWorkspace(), pureCallCases(t)
	manifest.Packages[1].Sources[0].Content += "activity Start(Integer) -> Integer computes `return input + 1`\nbind Start.result -> Main.input\n"
	suite.Cases = suite.Cases[:1]
	suite.Cases[0].Inputs = map[string]json.RawMessage{"example/app:Start": json.RawMessage(`3`)}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || len(result.Program.Activities) != 2 || result.Runtime.FinitePassed != 1 || len(result.Composition.Plan.Edges) != 1 {
		t.Fatal("explicit producer disappeared", err, result.Runtime)
	}
}

func TestWorkspacePureCallsRejectCyclesAfterResolution(t *testing.T) {
	m := pureCallWorkspace()
	m.Packages[0].Sources[1].Content = strings.Replace(m.Packages[0].Sources[1].Content, "return input * 3", "return Scale(input)", 1)
	_, err := ExecuteWorkspace(context.Background(), m, pureCallCases(t), "", "")
	if err == nil || !strings.Contains(err.Error(), "recursive pure activity call") {
		t.Fatal("cyclic imported closure was accepted", err)
	}
}

func TestWorkspacePureCallsUseSourceFileImportScope(t *testing.T) {
	m := pureCallWorkspace()
	m.Packages[1].Sources[0].Content = strings.Replace(m.Packages[1].Sources[0].Content, "import core \"example/core\"\n", "", 1)
	m.Packages[1].Sources = append(m.Packages[1].Sources, packageruntime.Source{Filename: "other.gooo",
		Content: "package app\nnamespace app\nimport core \"example/core\"\nactivity Elsewhere(Integer) -> Integer computes `return core.Scale(input)`\n"})
	if _, err := Prepare(m); err == nil {
		t.Fatal("another file's import alias became visible to Main")
	}
}

func TestWorkspacePureCallsDoNotCaptureParserWrapperName(t *testing.T) {
	m := pureCallWorkspace()
	source := &m.Packages[1].Sources[0].Content
	*source = strings.ReplaceAll(*source, "Double", "body")
	result, err := ExecuteWorkspace(context.Background(), m, pureCallCases(t), "", "")
	if err != nil || result.Runtime.FinitePassed != 2 {
		t.Fatal("parser wrapper captured a source activity name", err)
	}
}
