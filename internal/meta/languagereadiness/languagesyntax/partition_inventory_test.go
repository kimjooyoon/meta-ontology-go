package languagesyntax

import (
	"io/fs"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
)

const partitionDirectory = "internal/meta/executioncontinuation/policy"

func inventoryPartitionFixture(t *testing.T) (fstest.MapFS, []byte, sourcebundle.Partition) {
	t.Helper()
	header := "package policy\nnamespace policy\n"
	source := header + "entity Integer id \"urn:gooo:type:integer\"\n" +
		"# 계산의 원본\nactivity Twice(Integer) -> Integer id \"urn:twice\" computes `return input * 2`\n" +
		"activity Value(Integer) -> Integer id \"urn:value\" computes `return Twice(input)`\n"
	source = strings.ReplaceAll(source, "\n", "\r\n")
	files := fstest.MapFS{
		partitionDirectory + "/source.gooo":   {Data: []byte(source)},
		partitionDirectory + "/a/nested.gooo": {Data: []byte(header + "entity Neighbor id \"urn:neighbor\"\n")},
	}
	raw, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := sourcebundle.PreparePartition([]sourcebundle.Source{{Filename: "source.gooo", Content: source}},
		"source.gooo", "urn:value", "a.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return files, raw, proposal
}

func TestPartitionInventoryMatchesActualInstalledBytes(t *testing.T) {
	files, registry, proposal := inventoryPartitionFixture(t)
	original := cloneInventoryFiles(files)
	before, err := InspectSourceInventory(files, registry)
	if err != nil {
		t.Fatal(err)
	}
	result, err := InspectPartitionInventory(files, registry, partitionDirectory, proposal)
	if err != nil {
		t.Fatal(err)
	}
	want := partitionDirectory + "/a.gooo"
	if !reflect.DeepEqual(result.Before, before) || !reflect.DeepEqual(files, original) ||
		!reflect.DeepEqual(result.AddedUnregistered, []string{want}) ||
		!reflect.DeepEqual(result.Projected.Missing, before.Missing) || len(before.Missing) == 0 ||
		len(result.Projected.Unregistered) != len(before.Unregistered)+1 {
		t.Fatal("projection lost existing debt or changed source", err, result)
	}
	if !slices.Contains(result.Projected.Unregistered, partitionDirectory+"/a/nested.gooo") ||
		result.Projected.RegistryDigest != before.RegistryDigest || result.Projected.SemanticAdmission != "UNASSESSED" ||
		result.Projected.RepositoryWrites != 0 || result.Projected.NativeExecutions != 0 {
		t.Fatal("inventory claimed registration, admission, or execution", result)
	}
	installed := cloneInventoryFiles(files)
	for _, output := range proposal.Files {
		installed[path.Join(partitionDirectory, output.Filename)] = &fstest.MapFile{Data: []byte(output.Source)}
	}
	actual, err := InspectSourceInventory(installed, registry)
	if err != nil || !reflect.DeepEqual(result.Projected, actual) {
		t.Fatal("prediction differs from observing actual installed bytes", err, result.Projected, actual)
	}
	// WalkDir visits the a subtree before its sibling a.gooo. Sorting complete
	// paths instead would change the observation digest despite identical bytes.
	if result.Projected.Sources[0].Path != partitionDirectory+"/a/nested.gooo" ||
		result.Projected.Sources[1].Path != want {
		t.Fatal("directory traversal order changed", result.Projected.Sources)
	}
	second, err := InspectPartitionInventory(files, registry, partitionDirectory, proposal)
	if err != nil || !reflect.DeepEqual(result, second) {
		t.Fatal("partition inventory is not replayable", err)
	}
}

func TestPartitionInventoryRejectsStaleOrForgedPlans(t *testing.T) {
	mutations := []string{"stale", "occupied", "case-alias", "output", "origin", "meaning", "boundary",
		"directory", "registry", "empty"}
	for _, mutation := range mutations {
		t.Run(mutation, func(t *testing.T) {
			files, registry, proposal := inventoryPartitionFixture(t)
			directory := partitionDirectory
			switch mutation {
			case "stale":
				files[partitionDirectory+"/source.gooo"].Data = append(files[partitionDirectory+"/source.gooo"].Data, '#')
			case "occupied", "case-alias":
				name := "a.gooo"
				if mutation == "case-alias" {
					name = "A.gooo"
				}
				files[path.Join(directory, name)] = &fstest.MapFile{Data: []byte("occupied")}
			case "output":
				proposal.Files[1].Source += "# invented\n"
			case "origin":
				proposal.Files[1].Origins[0].InputStart++
			case "meaning":
				proposal.After.SourceDigest = strings.Repeat("0", 64)
			case "boundary":
				proposal.BoundaryID = "urn:twice"
			case "directory":
				directory = "../policy"
			case "registry":
				registry = []byte(`{}`)
			case "empty":
				proposal = sourcebundle.Partition{}
			}
			original := cloneInventoryFiles(files)
			result, err := InspectPartitionInventory(files, registry, directory, proposal)
			if err == nil || !reflect.DeepEqual(result, PartitionInventory{}) || !reflect.DeepEqual(files, original) {
				t.Fatal("invalid projection yielded a usable inventory or wrote source", mutation, err)
			}
		})
	}
}

func cloneInventoryFiles(files fstest.MapFS) fstest.MapFS {
	copy := make(fstest.MapFS, len(files))
	for name, file := range files {
		cloned := *file
		cloned.Data = slices.Clone(file.Data)
		copy[name] = &cloned
	}
	return copy
}

type changingPartitionFS struct {
	fs.FS
	files       fstest.MapFS
	reads       int
	destination bool
}

func (f *changingPartitionFS) Open(name string) (fs.File, error) {
	if name == partitionDirectory+"/source.gooo" {
		f.reads++
		if f.reads == 3 {
			if f.destination {
				f.files[partitionDirectory+"/a.gooo"] = &fstest.MapFile{Data: []byte("occupied after observation")}
			} else {
				f.files[name].Data = append(f.files[name].Data, []byte("# changed during observation\n")...)
			}
		}
	}
	return f.FS.Open(name)
}

func TestPartitionInventoryRechecksSourcesAfterObservation(t *testing.T) {
	for _, destination := range []bool{false, true} {
		files, registry, proposal := inventoryPartitionFixture(t)
		changing := &changingPartitionFS{FS: files, files: files, destination: destination}
		result, err := InspectPartitionInventory(changing, registry, partitionDirectory, proposal)
		if err == nil || !reflect.DeepEqual(result, PartitionInventory{}) || changing.reads < 3 {
			t.Fatal("source or destination changed during observation but prediction was retained", err, changing.reads)
		}
	}
}

func TestPartitionInventoryCanSupplyAnAlreadyRegisteredMissingMember(t *testing.T) {
	files, registry, previous := inventoryPartitionFixture(t)
	sources := []sourcebundle.Source{{Filename: "source.gooo", Content: string(files[partitionDirectory+"/source.gooo"].Data)}}
	proposal, err := sourcebundle.PreparePartition(sources, "source.gooo", previous.BoundaryID, "source_recovery.gooo")
	if err != nil {
		t.Fatal(err)
	}
	result, err := InspectPartitionInventory(files, registry, partitionDirectory, proposal)
	if err != nil {
		t.Fatal(err)
	}
	name := partitionDirectory + "/source_recovery.gooo"
	if !slices.Contains(result.Before.Missing, name) || slices.Contains(result.Projected.Missing, name) ||
		len(result.AddedUnregistered) != 0 || !reflect.DeepEqual(result.Before.Unregistered, result.Projected.Unregistered) {
		t.Fatal("an already registered proposed member must resolve only its existing missing path", result)
	}
}
