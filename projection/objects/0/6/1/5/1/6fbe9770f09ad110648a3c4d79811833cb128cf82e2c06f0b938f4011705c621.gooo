package cycles

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type fixture struct {
	Name     string       `json:"name"`
	Graph    Graph        `json:"graph"`
	Expected map[Code]int `json:"expected"`
}

func TestFixturesMatchExpectedDiagnostics(t *testing.T) {
	for _, name := range []string{"minimal-invalid.json", "cross-namespace-negative.json"} {
		current := loadFixture(t, name)
		diagnostics := Detect(current.Graph)
		counts := make(map[Code]int)
		for _, diagnostic := range diagnostics {
			counts[diagnostic.Code]++
		}
		if !reflect.DeepEqual(counts, current.Expected) {
			t.Fatalf("fixture %q expected %#v, got %#v (%v)", name, current.Expected, counts, diagnostics)
		}
	}
}
func TestFixtureDiagnosticsAreStableAcrossInputOrder(t *testing.T) {
	fixture := loadFixture(t, "minimal-invalid.json")
	reversed := fixture.Graph
	reversed.Nodes = reverseNodes(fixture.Graph.Nodes)
	reversed.Edges = reverseEdges(fixture.Graph.Edges)
	left, right := Detect(fixture.Graph), Detect(reversed)
	if !reflect.DeepEqual(left, right) || left.Error() != right.Error() {
		t.Fatalf("fixture diagnostics changed with insertion order:\nleft=%v\nright=%v", left, right)
	}
}
func loadFixture(t *testing.T, name string) fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var current fixture
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&current); err != nil {
		t.Fatalf("decode fixture %q: %v", name, err)
	}
	if current.Name == "" {
		t.Fatalf("fixture %q has no name", name)
	}
	for code, count := range current.Expected {
		if count < 0 {
			t.Fatalf("fixture %q has negative count for %q", name, code)
		}
	}
	return current
}
func reverseNodes(nodes []Node) []Node {
	result := append([]Node(nil), nodes...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
