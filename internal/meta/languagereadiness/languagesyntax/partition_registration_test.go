package languagesyntax

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"reflect"
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcebundle"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func registeredPartition(t *testing.T, directory string, names []string, source, before, into string,
	profiles ...syntax.EntityFieldsSupport) (fs.FS, []byte, sourcebundle.Partition) {
	t.Helper()
	repository := os.DirFS("../../../..")
	raw, err := fs.ReadFile(repository, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	files, err := fs.Sub(repository, directory)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := sourcebundle.Read(files, names)
	if err != nil {
		t.Fatal(err)
	}
	var partition sourcebundle.Partition
	if len(profiles) == 0 {
		partition, err = sourcebundle.PreparePartition(sources, source, before, into)
	} else {
		if len(profiles) != 1 {
			t.Fatal("test requires exactly one explicit profile")
		}
		partition, err = sourcebundle.PreparePartitionWithEntityFieldsSupport(sources, source, before, into, profiles[0])
	}
	if err != nil {
		t.Fatal(err)
	}
	return repository, raw, partition
}

func TestRegistrationPartitionKeepsOneCapabilityAndAllItsActivities(t *testing.T) {
	repository, raw, partition := registeredPartition(t, partitionDirectory,
		[]string{"source.gooo", "source_recovery.gooo"}, "source_recovery.gooo",
		"gooo://self-improvement/consumer/next-observation", "source_observation.gooo")
	result, err := PreparePartitionRegistration(repository, raw, partitionDirectory, partition)
	if err != nil {
		t.Fatal(err)
	}
	before, err := decodeRegistry(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Registry.Cases, before.Cases) || len(result.Registry.CaseSources) != 1 ||
		len(result.Registry.CaseSources[0].Members) != 3 || len(result.Consumers) != 1 ||
		len(result.Consumers[0].Before.SourceMembers) != 2 || len(result.Consumers[0].After.SourceMembers) != 3 ||
		result.Consumers[0].Before.SemanticDigest != result.Consumers[0].After.SemanticDigest {
		t.Fatal("partition changed capability identity or omitted part of the family", result.Consumers)
	}
	newPath := path.Join(partitionDirectory, "source_observation.gooo")
	if !slices.Contains(result.Registry.MetaSources, newPath) || len(result.Inventory.AddedUnregistered) != 0 ||
		len(result.Inventory.Projected.Missing) != 0 || len(result.Inventory.Projected.Unregistered) != 0 ||
		len(result.Inventory.Projected.Sources) != len(result.Inventory.Before.Sources)+1 ||
		result.Inventory.Projected.RepositoryWrites != 0 || result.Inventory.Projected.NativeExecutions != 0 ||
		result.Inventory.Projected.SemanticAdmission != "UNASSESSED" ||
		result.Inventory.Projected.RegistryDigest != digestBytes(result.Corpus) {
		t.Fatal("physical registration or planning authority diverged")
	}
	var decoded Registry
	if err := json.Unmarshal(result.Corpus, &decoded); err != nil || !reflect.DeepEqual(decoded, result.Registry) {
		t.Fatal("projected corpus bytes differ from the planned registry", err)
	}
	second, err := PreparePartitionRegistration(repository, raw, partitionDirectory, partition)
	if err != nil || !reflect.DeepEqual(result, second) {
		t.Fatal("registration projection is not deterministic", err)
	}
}

func TestRegistrationPartitionPreservesMetaSourceAndPackageRoles(t *testing.T) {
	for _, item := range []struct {
		directory, source, boundary, destination string
		names                                    []string
		kind                                     string
		support                                  syntax.EntityFieldsSupport
	}{
		{"examples/agent-tool-inventory", "main.gooo", "urn:gooo:agent-tool-inventory:inspect", "inspect.gooo", []string{"main.gooo"}, "meta", syntax.EntityFieldsV5Support()},
		{"examples/billing-package", "entities.gooo", "urn:gooo:billing:receipt", "receipt.gooo", []string{"activity.gooo", "entities.gooo"}, "package", syntax.CurrentEntityFieldsSupport()},
	} {
		t.Run(item.kind, func(t *testing.T) {
			repository, raw, partition := registeredPartition(t, item.directory, item.names, item.source, item.boundary,
				item.destination, item.support)
			before, _ := decodeRegistry(raw)
			result, err := PreparePartitionRegistration(repository, raw, item.directory, partition)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Registry.Cases, before.Cases) ||
				!reflect.DeepEqual(result.Registry.CaseSources, before.CaseSources) || len(result.Inventory.AddedUnregistered) != 0 {
				t.Fatal("physical source member created a language capability")
			}
			if item.kind == "meta" {
				if len(result.Consumers) != 0 || len(result.Registry.MetaSources) != len(before.MetaSources)+1 {
					t.Fatal("meta source was promoted into a syntax case")
				}
			} else if len(result.Consumers) != 1 || result.Consumers[0].Kind != "package-syntax" ||
				len(result.Registry.PackageUnits[0].Members) != 3 ||
				result.Registry.PackageUnits[0].Entry != before.PackageUnits[0].Entry {
				t.Fatal("package entry or syntax-only observation changed")
			}
		})
	}
}
