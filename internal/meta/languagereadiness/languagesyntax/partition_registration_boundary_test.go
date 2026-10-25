package languagesyntax

import (
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRegistrationPartitionStartsAGroupWithoutChangingTheCase(t *testing.T) {
	const directory = "examples/entity-fields-v1"
	repository := os.DirFS("../../../..")
	raw, err := fs.ReadFile(repository, directory+"/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sourcebundle.LowerWithEntityFieldsSupport(
		[]sourcebundle.Source{{Filename: "main.gooo", Content: string(raw)}}, syntax.EntityFieldsV1Support())
	if err != nil {
		t.Fatal(err)
	}
	repositoryFS, registry, partition := registeredPartition(t, directory, []string{"main.gooo"}, "main.gooo",
		bundle.Provenance.Declarations[1].ID, "activity.gooo", syntax.EntityFieldsV1Support())
	result, err := PreparePartitionRegistration(repositoryFS, registry, directory, partition)
	if err != nil {
		t.Fatal(err)
	}
	names := caseSourcesFor(result.Registry, "entity-fields-v1")
	if !reflect.DeepEqual(names, []string{directory + "/main.gooo", directory + "/activity.gooo"}) ||
		len(result.Consumers) != 1 || result.Consumers[0].Before.SourceDigestKind != "" ||
		result.Consumers[0].After.SourceDigestKind != "gooo/source-bundle/v1" {
		t.Fatal("single-file case was replaced instead of extending physical membership")
	}
	// The partitioner can read integer fields, but this registered case selected
	// V1. Its consumer must not silently widen that explicit language profile.
	changed := strings.ReplaceAll(string(raw), "type string", "type integer")
	altered := sourcebundle.Overlay{FS: repository, Files: map[string][]byte{directory + "/main.gooo": []byte(changed)}}
	proposal, err := sourcebundle.PreparePartitionWithEntityFieldsSupport([]sourcebundle.Source{{Filename: "main.gooo", Content: changed}},
		"main.gooo", bundle.Provenance.Declarations[1].ID, "activity.gooo", syntax.EntityFieldsV3Support())
	if err != nil {
		t.Fatal(err)
	}
	if result, err := PreparePartitionRegistration(altered, registry, directory, proposal); err == nil ||
		!reflect.DeepEqual(result, PartitionRegistration{}) {
		t.Fatal("consumer profile was widened to fit a partition", err)
	}
}

func TestRegistrationPartitionRetainsUnrelatedInventoryDebt(t *testing.T) {
	const directory = "examples/agent-tool-inventory"
	repository, registry, partition := registeredPartition(t, directory, []string{"main.gooo"}, "main.gooo",
		"urn:gooo:agent-tool-inventory:inspect", "inspect.gooo", syntax.EntityFieldsV5Support())
	raw, err := fs.ReadFile(repository, directory+"/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{directory + "/main.gooo": {Data: raw},
		"examples/unregistered/main.gooo": {Data: []byte("package p\nnamespace p\nentity X\n")}}
	result, err := PreparePartitionRegistration(files, registry, directory, partition)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Inventory.Before.Missing) == 0 ||
		!reflect.DeepEqual(result.Inventory.Projected.Missing, result.Inventory.Before.Missing) ||
		!reflect.DeepEqual(result.Inventory.Projected.Unregistered, result.Inventory.Before.Unregistered) {
		t.Fatal("unrelated missing or unregistered files disappeared")
	}
}

type changingRegistrySibling struct {
	fs.FS
	name            string
	changeAt, reads int
	replacement     []byte
}

func (f *changingRegistrySibling) Open(name string) (fs.File, error) {
	if name == f.name {
		f.reads++
		if f.reads == f.changeAt {
			return (sourcebundle.Overlay{Files: map[string][]byte{name: f.replacement}}).Open(name)
		}
	}
	return f.FS.Open(name)
}

func TestRegistrationPartitionPinsAnUnchangedSiblingOutsideTheSelectedSplit(t *testing.T) {
	repository, registry, partition := registeredPartition(t, partitionDirectory, []string{"source.gooo"},
		"source.gooo", "gooo://self-improvement/consumer/workflow-path", "source_workflow.gooo")
	sibling := partitionDirectory + "/source_recovery.gooo"
	raw, err := fs.ReadFile(repository, sibling)
	if err != nil {
		t.Fatal(err)
	}
	for _, changeAt := range []int{2, 3} {
		files := &changingRegistrySibling{FS: repository, name: sibling, changeAt: changeAt,
			replacement: append(raw, []byte("\n// different origin\n")...)}
		result, err := PreparePartitionRegistration(files, registry, partitionDirectory, partition)
		if err == nil || !reflect.DeepEqual(result, PartitionRegistration{}) || files.reads < changeAt {
			t.Fatal("changed sibling yielded a usable registry projection", changeAt, err, files.reads)
		}
	}
}

func TestRegistrationPartitionDoesNotReplaceExistingOrUnsupportedRoles(t *testing.T) {
	registry := expectedRegistry()
	for _, pair := range [][2]string{
		{"examples/agent-tool-inventory/main.gooo", "examples/entity-fields-v1/main.gooo"},
		{"examples/agent-tool-inventory/main.gooo", "examples/entity-fields-v1/MAIN.gooo"},
		{"examples/agent-tool-inventory/MAIN.gooo", "examples/new.gooo"},
		{"examples/unregistered/main.gooo", "examples/new.gooo"},
		{"internal/completeness/receipt.gooo", "internal/completeness/members.gooo"},
	} {
		if err := partitionRegistrationAddresses(registry, pair[0], pair[1]); err == nil {
			t.Fatal("a source role was silently replaced or omitted", pair)
		}
	}
}
