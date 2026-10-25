package workspaceinput_test

import (
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestFailedWorkspaceLoadCannotBecomeExecutableSubset(t *testing.T) {
	definition := workspaceinput.Definition{Schema: workspaceinput.Schema,
		Entry:    packageruntime.EntrySpec{PackagePath: "first", Activity: "Keep"},
		Packages: []workspaceinput.Package{{Path: "first", Sources: []string{"first.gooo"}}, {Path: "later", Sources: []string{"missing.gooo"}}}}
	files := fstest.MapFS{"first.gooo": {Data: []byte("package first\nnamespace first\nentity Integer id \"urn:gooo:type:integer\"\nactivity Keep(Integer) -> Integer computes \"return input\"\n")}}
	partial, err := workspaceinput.Load(files, "workspace.json", definition)
	if err == nil {
		t.Fatal("missing input accepted")
	}
	if _, semanticErr := packageruntime.Describe(partial); semanticErr == nil {
		t.Fatal("failed source-set load returned an executable subset of the requested workspace")
	}
}
