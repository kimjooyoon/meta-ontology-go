package workspacesplit_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
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

func TestJointDefinitionCannotOverwriteAnotherObservedSource(t *testing.T) {
	files, owners := sharedOwners()
	files["dependency.gooo"] = &fstest.MapFile{Data: []byte("package shared\nnamespace shared\n")}
	owners[0].Definition.Packages[0].Sources = append(owners[0].Definition.Packages[0].Sources, "dependency.gooo")
	owners[1].ManifestPath = "dependency.gooo"
	plan, err := workspacesplit.Prepare(files, owners)
	if err == nil || !strings.Contains(err.Error(), "would overwrite an observed source") ||
		!reflect.DeepEqual(plan, workspacesplit.Plan{}) {
		t.Fatal("definition address lost its dependency consumer", err)
	}
}

func TestJointObservationsRequireAnUnderlyingReader(t *testing.T) {
	_, owners := sharedOwners()
	plan, err := workspacesplit.Prepare(nil, owners)
	if err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) {
		t.Fatal("missing reader produced a usable proposal", err)
	}
}
