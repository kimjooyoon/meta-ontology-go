package entityfieldsv1

import (
	"strings"
	"testing"
)

func TestObserveCanonicalEntityFieldsV1(t *testing.T) {
	observation, err := Observe("fixture.gooo", CanonicalSource)
	if err != nil { t.Fatal(err) }
	if observation.SourceDigest == "" || observation.FormattedDigest == "" || observation.GeneratedDigest == "" || observation.SourceMapDigest == "" { t.Fatal("missing source-backed digests") }
	if len(observation.Semantic.Entities) != 1 || len(observation.Semantic.Entities[0].Fields) != 2 || len(observation.Semantic.Activities) != 1 { t.Fatalf("semantic shape = %#v", observation.Semantic) }
	if len(observation.DeclarationOrder) != 2 || observation.DeclarationOrder[0] != SourceID { t.Fatalf("declaration order = %#v", observation.DeclarationOrder) }
	if observation.Semantic.Activities[0].Name != "LoadOrder" { t.Fatalf("semantic activity order = %#v", observation.Semantic.Activities) }
	for _, field := range observation.Semantic.Entities[0].Fields {
		if field.IDSpan.Start.Offset == 0 || field.NameSpan.Start.Offset == 0 || field.TypeRefSpan.Start.Offset == 0 || field.PresenceSpan.Start.Offset == 0 || field.CardinalitySpan.Start.Offset == 0 { t.Fatalf("field source spans = %#v", field) }
	}
	if !observation.GetPutRoundTrip || !observation.PutGetRoundTrip || len(observation.SourceMap.Mappings) < 2 { t.Fatal("atomic BX or source-map evidence missing") }
	if !hasNavigationSymbol(observation.Symbols, "OrderNumber", FieldID) || !hasNavigationSymbol(observation.Symbols, "CustomerName", SecondFieldID) { t.Fatalf("field symbol missing: %#v", observation.Symbols) }
	if len(observation.StableIDs) != 4 || len(observation.Counterexamples) != 6 { t.Fatalf("stable identity/counterexample count = %d/%d", len(observation.StableIDs), len(observation.Counterexamples)) }
}

func hasNavigationSymbol(symbols []NavigationSymbol, name, id string) bool {
	for _, symbol := range symbols {
		if symbol.Name == name && symbol.ID == id { return true }
	}
	return false
}

func TestUnsupportedEntityFieldShapeProducesNoProjection(t *testing.T) {
	for _, testCase := range []struct{ name, replacement string }{
		{"type", "type number required one"}, {"presence", "type string optional one"}, {"cardinality", "type string required many"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			source := strings.Replace(CanonicalSource, "type string required one", testCase.replacement, 1)
			observation, err := Observe("fixture.gooo", source)
			if err == nil || len(observation.Generated) != 0 || len(observation.SourceMap.Mappings) != 0 { t.Fatalf("unsupported field was partially projected: err=%v observation=%#v", err, observation) }
		})
	}
}
