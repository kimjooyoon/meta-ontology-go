package metricprogram

import (
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/syntaxregistration"
)

func TestRegistryMembershipFresh1641(t *testing.T) {
	inventory, err := syntaxregistration.InspectInventory(os.DirFS("../../.."))
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Unregistered) != 0 || len(inventory.Missing) != 0 ||
		inventory.NativeExecutions != 0 || inventory.RepositoryWrites != 0 {
		t.Fatal("source inventory is incomplete or executed a native program")
	}
	origin := RegistrySourceOrigin()
	for _, item := range inventory.Sources {
		if item.Path == origin.SourcePath && item.SourceDigest == origin.SourceDigest {
			return
		}
	}
	t.Fatal("registry declaration is missing from the source inventory")
}

func TestRegistryBoundsFresh1643(t *testing.T) {
	for _, item := range []struct {
		name  string
		value any
	}{
		{"operation-before-first", OperationAt(-1)},
		{"operation-after-last", OperationAt(OperationCount(0))},
		{"entity-before-first", EntityAt(-1)},
		{"entity-after-last", EntityAt(EntityCount(0))},
	} {
		t.Run(item.name, func(t *testing.T) {
			if !reflect.ValueOf(item.value).IsZero() {
				t.Fatal("out-of-range registry lookup produced an operation or identity")
			}
		})
	}
}
