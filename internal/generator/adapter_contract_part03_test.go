package generator

import (
	"reflect"
	"strings"
	"testing"
)

func TestReflectiveFallbackPreservesAuthoritativeFactOrder(t *testing.T) {
	input := reflectiveGraph()
	input.Facts = []reflectiveFactFixture{
		{Subject: "activity:run", Predicate: "used", Object: "entity:z"},
		{Subject: "activity:run", Predicate: "used", Object: "entity:a"},
	}
	source, _, err := GenerateFrom(input, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "func Run(zeta Zeta, alpha Alpha)") {
		t.Fatalf("reflective slice fact order was not preserved:\n%s", source)
	}
}
func TestGenerateFromProjectionV1ReflectiveInputIsDeterministic(t *testing.T) {
	input := reflectiveGraph()
	first, err := GenerateFromProjectionV1(input, Options{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateFromProjectionV1(input, Options{})
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := first.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := second.CanonicalJSON()
	if err != nil || !reflect.DeepEqual(firstJSON, secondJSON) {
		t.Fatalf("adapter projection is not repeat-stable: %v", err)
	}
	if first.Metadata.Provenance.Status != "DEFERRED" || first.Metadata.Evidence.Decision != "DEFERRED" {
		t.Fatalf("external status was fabricated: %#v", first.Metadata)
	}
}
func TestGenerateFromProjectionV1RejectsMalformedInputWithoutMutation(t *testing.T) {
	input := reflectiveGraph()
	input.Facts = []reflectiveFactFixture{{Subject: "activity:run", Predicate: "used", Object: "entity:missing"}}
	before := input
	if _, err := GenerateFromProjectionV1(&input, Options{}); err == nil {
		t.Fatal("malformed reflective input was accepted")
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("adapter rejection mutated caller input")
	}
}
func reflectiveGraph() reflectiveGraphFixture {
	return reflectiveGraphFixture{Package: "reflectgen", Nodes: baseNodes(), Facts: []reflectiveFactFixture{}}
}
func baseNodes() map[string]reflectiveNodeFixture {
	return map[string]reflectiveNodeFixture{
		"activity": {ID: "activity:run", Kind: "activity", Name: "Run"},
		"alpha":    {ID: "entity:a", Kind: "entity", Name: "Alpha"},
		"zeta":     {ID: "entity:z", Kind: "entity", Name: "Zeta"},
	}
}
func withNode(input reflectiveGraphFixture, node reflectiveNodeFixture) reflectiveGraphFixture {
	input.Nodes = map[string]reflectiveNodeFixture{"base": node}
	return input
}
func withNodes(input reflectiveGraphFixture, nodes any) reflectiveGraphFixture {
	input.Nodes = nodes
	return input
}
