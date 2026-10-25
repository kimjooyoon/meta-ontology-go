package languagesyntax

import (
	"io/fs"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

func TestPartitionInventoryTracksActualContinuationFamily(t *testing.T) {
	repository := os.DirFS("../../../..")
	registry, err := fs.ReadFile(repository, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	family, err := fs.Sub(repository, partitionDirectory)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := sourcebundle.Read(family, []string{"source.gooo", "source_recovery.gooo"})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := sourcebundle.PreparePartition(sources, "source_recovery.gooo",
		"gooo://self-improvement/consumer/next-observation", "source_observation.gooo")
	if err != nil {
		t.Fatal(err)
	}
	result, err := InspectPartitionInventory(repository, registry, partitionDirectory, proposal)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Before.Missing) != 0 || len(result.Before.Unregistered) != 0 ||
		len(result.Projected.Missing) != 0 || !reflect.DeepEqual(result.Projected.Unregistered,
		[]string{partitionDirectory + "/source_observation.gooo"}) {
		t.Fatal("actual native family must expose precisely the new registration need", result)
	}
	installed := fstest.MapFS{}
	for _, source := range result.Before.Sources {
		raw, err := fs.ReadFile(repository, source.Path)
		if err != nil {
			t.Fatal(err)
		}
		installed[source.Path] = &fstest.MapFile{Data: raw}
	}
	for _, output := range proposal.Files {
		installed[partitionDirectory+"/"+output.Filename] = &fstest.MapFile{Data: []byte(output.Source)}
	}
	actual, err := InspectSourceInventory(installed, registry)
	if err != nil || !reflect.DeepEqual(result.Projected, actual) {
		t.Fatal("real source prediction differs from the installed in-memory source tree", err)
	}
}
