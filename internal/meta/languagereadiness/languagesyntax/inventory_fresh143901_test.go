package languagesyntax

import (
	"os"
	"testing"
	"testing/fstest"
)

func TestSourceInventoryWithoutReplay143901(t *testing.T) {
	root := os.DirFS("../../../..")
	raw, err := os.ReadFile("../../../../examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	current, err := InspectSourceInventory(root, raw)
	if err != nil || len(current.Unregistered) != 0 || len(current.Missing) != 0 || current.SemanticAdmission != "UNASSESSED" || current.NativeExecutions != 0 || current.RepositoryWrites != 0 {
		t.Fatal(current, err)
	}
	fixture := fstest.MapFS{"application/fresh143901.gooo": {Data: []byte("package fresh143901\nnamespace fresh143901\n")}}
	next, err := InspectSourceInventory(fixture, raw)
	if err != nil || len(next.Unregistered) != 1 || next.Unregistered[0] != "application/fresh143901.gooo" || len(next.Missing) == 0 || next.ObservationDigest == current.ObservationDigest {
		t.Fatal(next, err)
	}
}
