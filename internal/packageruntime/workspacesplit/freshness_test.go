package workspacesplit_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

type changingReader struct{ read func(string) ([]byte, error) }

func (r changingReader) ReadFile(path string) ([]byte, error) { return r.read(path) }

func TestJointSplitRefusesChangedSharedBytesAcrossOwners(t *testing.T) {
	files, owners := sharedOwners()
	reads := 0
	reader := changingReader{read: func(path string) ([]byte, error) {
		raw, err := files.ReadFile(path)
		reads++
		if reads == 2 {
			raw = []byte(strings.Replace(string(raw), "input+7", "input+8", 1))
		}
		return raw, err
	}}
	plan, err := workspacesplit.Prepare(reader, owners)
	if err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) || reads != 2 {
		t.Fatal("inconsistent source observations produced a joint draft", plan, err, reads)
	}
}

func TestJointSplitCanonicalizesAddressesAndRetainsAuthoredSourceAlias(t *testing.T) {
	files, owners := sharedOwners()
	owners[0].Definition.Packages[0].Sources = []string{"./shared.gooo"}
	owners[1].ManifestPath = "nested/../second.workspace.json"
	owners[1].Request.Source = "nested/../shared.gooo"
	plan, err := workspacesplit.Prepare(files, owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Files) != 2 || plan.Owners[0].Split.Definition.Packages[0].Sources[0] != "./shared.gooo" {
		t.Fatal("lexical aliases duplicated the shared address or rewrote authored spelling", plan)
	}
}
