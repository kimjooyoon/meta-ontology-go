package cycles

import (
	"os"
	"reflect"
	"testing"
)

func TestResearchFixturesEvaluateAgainstExactCounts(t *testing.T) {
	for _, name := range []string{"minimal-invalid.json", "cross-namespace-negative.json"} {
		fixture := loadResearchFixture(t, name)
		result := fixture.Evaluate()
		if result.Outcome != OutcomePass {
			t.Fatalf("fixture %q failed: %#v", name, result)
		}
		if len(result.Measure.Digest) != 64 || len(result.Deferred) == 0 {
			t.Fatalf("fixture %q omitted evidence or deferred work: %#v", name, result)
		}
	}
}

func TestMinimalFixtureMeasuresEachDefect(t *testing.T) {
	fixture := loadResearchFixture(t, "minimal-invalid.json")
	measurement := Measure(fixture.Graph)
	if measurement.NodeCount != 3 || measurement.EdgeCount != 4 || measurement.DiagnosticCount != 4 {
		t.Fatalf("unexpected minimal fixture measurement: %#v", measurement)
	}
	for _, code := range []Code{CycleDetected, IllegalRelationDirection, NamespaceCollision, UnresolvedStableID} {
		if measurement.CodeCounts[code] != 1 {
			t.Fatalf("expected one %q diagnostic: %#v", code, measurement)
		}
	}
}

func TestFixtureMeasurementIsInputOrderIndependent(t *testing.T) {
	fixture := loadResearchFixture(t, "minimal-invalid.json")
	reversed := fixture.Graph
	reversed.Nodes = reverseNodes(fixture.Graph.Nodes)
	reversed.Edges = reverseEdges(fixture.Graph.Edges)
	left, right := Measure(fixture.Graph), Measure(reversed)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("measurement changed with insertion order:\nleft=%#v\nright=%#v", left, right)
	}
}

func TestFixtureValidationRequiresDeferredFutureWork(t *testing.T) {
	fixture := loadResearchFixture(t, "cross-namespace-negative.json")
	fixture.FollowUp.Deferred = nil
	if err := fixture.Validate(); err == nil {
		t.Fatal("fixture without deferred future work was accepted")
	}
}

func loadResearchFixture(t *testing.T, name string) ResearchFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := LoadFixture(data)
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func reverseNodes(nodes []Node) []Node {
	result := append([]Node(nil), nodes...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseEdges(edges []Edge) []Edge {
	result := append([]Edge(nil), edges...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
