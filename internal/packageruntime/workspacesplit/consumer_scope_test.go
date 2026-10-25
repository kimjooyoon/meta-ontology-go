package workspacesplit_test

import (
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
)

func TestOneWorkspaceProposalCannotDischargeAnotherSharedConsumer(t *testing.T) {
	files, definition, request := sharedFixture()
	for _, path := range []string{"first.workspace.json", "second.workspace.json"} {
		loaded, err := workspaceinput.Load(files, path, definition)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := packageruntime.Describe(loaded); err != nil {
			t.Fatal(err)
		}
	}
	proposal, err := workspaceinput.PrepareSplit(files, "first.workspace.json", definition, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range proposal.Result.Files {
		files[file.Filename] = &fstest.MapFile{Data: []byte(file.Source)}
	}
	loaded, err := workspaceinput.Load(files, "second.workspace.json", definition)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := packageruntime.Describe(loaded); err == nil {
		t.Fatal("an unchanged shared consumer unexpectedly retained the moved entry")
	}
}
