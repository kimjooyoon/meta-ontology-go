package workspaceinput_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestNativeConsumerLoadsSplitsAndExecutesWorkspaceWithoutCLI(t *testing.T) {
	definition, err := workspaceinput.Decode([]byte(`{
 "schema":"gooo/package-workspace-manifest/v1",
 "entry":{"package_path":"shared/app","activity":"Run"},
 "packages":[{"path":"shared/math","sources":["math.gooo"]},{"path":"shared/app","sources":["app.gooo"]}]
}`))
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{
		"math.gooo": {Data: []byte("package math\nnamespace math\nentity Integer id \"urn:gooo:type:integer\"\nactivity Add(Integer) -> Integer id \"urn:shared:add\" computes \"return input+29\"\n")},
		"app.gooo":  {Data: []byte("package app\nnamespace app\nimport math \"shared/math\"\nactivity Credit(math.Integer) -> math.Integer id \"urn:shared:credit\" computes \"return input+7\"\nactivity Run(math.Integer) -> math.Integer id \"urn:shared:run\" computes \"return math.Add(input)\"\nbind Credit.result -> Run.input\n")},
	}
	loaded, err := workspaceinput.Load(files, "workspace.json", definition)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Packages[1].Name != "app" || len(loaded.Packages[1].Imports) != 1 || loaded.Packages[1].Imports[0] != "shared/math" {
		t.Fatal("source metadata was not derived for the native consumer")
	}
	if definition.Packages[1].Name != "" || definition.Packages[1].Imports != nil {
		t.Fatal("derived metadata rewrote the authored external definition")
	}
	proposal, err := workspaceinput.PrepareSplit(files, "workspace.json", definition, workspaceinput.SplitRequest{
		PackagePath: "shared/app", Source: "app.gooo", BeforeID: "urn:shared:run", Into: "run.gooo"})
	if err != nil {
		t.Fatal(err)
	}
	if len(definition.Packages[1].Sources) != 1 || proposal.Definition.Packages[1].Name != "" ||
		proposal.Definition.Packages[1].Imports != nil {
		t.Fatal("native source plan changed the authored definition")
	}
	for _, file := range proposal.Result.Files {
		files[file.Filename] = &fstest.MapFile{Data: []byte(file.Source)}
	}
	afterBytes, err := json.Marshal(proposal.Definition)
	if err != nil {
		t.Fatal(err)
	}
	afterDefinition, err := workspaceinput.Decode(afterBytes)
	if err != nil {
		t.Fatal(err)
	}
	after, err := workspaceinput.Load(files, "workspace.json", afterDefinition)
	if err != nil {
		t.Fatal(err)
	}
	view, err := packageruntime.Describe(after)
	if err != nil || view.ImageDigest != proposal.Result.OutputImageDigest {
		t.Fatal("authored output did not reconstruct the planned image", view.ImageDigest, err)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"shared/app:Credit":5},"expected":{"shared/app:Credit":12,"shared/app:Run":41}},
 {"inputs":{"shared/app:Credit":-7},"expected":{"shared/app:Credit":0,"shared/app:Run":29}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	for name, manifest := range map[string]packageruntime.Manifest{"loaded": loaded, "reloaded proposal": after} {
		t.Run(name, func(t *testing.T) {
			result, err := workspaceexecution.ExecuteWorkspace(context.Background(), manifest, suite, "", "")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Runtime
			if r.FinitePassed != 4 || r.FiniteTotal != 4 || !r.RuntimeReplayed || r.ModelCalls != 0 {
				t.Fatal("native consumer output/replay obligations", r)
			}
		})
	}
}

func TestNativeConsumerRejectsUnvalidatedDefinitionAndOversizeSource(t *testing.T) {
	definition := workspaceinput.Definition{Schema: "wrong-version", Packages: []workspaceinput.Package{{Path: "p", Sources: []string{"p.gooo"}}}}
	if _, err := workspaceinput.Load(fstest.MapFS{}, "workspace.json", definition); err == nil || !strings.Contains(err.Error(), "schema must be") {
		t.Fatal("native caller bypassed the external definition contract", err)
	}
	definition.Schema = workspaceinput.Schema
	files := fstest.MapFS{"p.gooo": {Data: make([]byte, workspaceinput.MaxSourceBytes+1)}}
	if _, err := workspaceinput.Load(files, "workspace.json", definition); err == nil || !strings.Contains(err.Error(), "input exceeds maximum size") {
		t.Fatal("native caller bypassed the source byte limit", err)
	}
}
