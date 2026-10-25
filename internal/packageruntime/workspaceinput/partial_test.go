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

type sourceReadFunc func(string) ([]byte, error)

func (r sourceReadFunc) ReadFile(path string) ([]byte, error) { return r(path) }

func TestWorkspaceSharesOneSourceObservationOnlyWithinOneLoad(t *testing.T) {
	definition, files, _ := nativeSplitFixture()
	definition.Packages = append(definition.Packages, workspaceinput.Package{
		Path: "native/other", Sources: []string{"./main.gooo"}})
	reads := 0
	reader := sourceReadFunc(func(path string) ([]byte, error) { reads++; return files.ReadFile(path) })
	first, err := workspaceinput.Load(reader, "workspace.json", definition)
	if err != nil || reads != 1 || first.Packages[0].Sources[0].Content != first.Packages[1].Sources[0].Content {
		t.Fatal("shared source was observed separately within one definition", reads, err)
	}
	files["main.gooo"].Data = append(files["main.gooo"].Data, []byte("\n# later source revision\n")...)
	second, err := workspaceinput.Load(reader, "workspace.json", definition)
	if err != nil || reads != 2 || first.Packages[0].Sources[0].Content == second.Packages[0].Sources[0].Content ||
		second.Packages[1].Sources[0].Content != string(files["main.gooo"].Data) {
		t.Fatal("source observation was cached across separate loads", reads, err)
	}
}
