package workspaceinput_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestNativeConsumerLoadsSplitsAndExecutesWorkspaceWithoutCLI(t *testing.T) {
	definition, files, loaded := loadImportedNativeFixture(t)
	after := reloadNativeSplitFixture(t, files, definition)
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

func loadImportedNativeFixture(t *testing.T) (workspaceinput.Definition, fstest.MapFS, packageruntime.Manifest) {
	t.Helper()
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
	return definition, files, loaded
}
