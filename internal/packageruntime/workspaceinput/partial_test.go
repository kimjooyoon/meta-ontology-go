package workspaceinput_test

import (
	"strings"
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
