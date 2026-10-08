package syntax

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSearchAlternativesRoundTripAndOwnedStorage(t *testing.T) {
	raw, err := os.ReadFile("../../examples/search-feedback/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := Parse(string(raw))
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	before := file.Declarations[1].(*ActivityDecl).Assembly.Spec
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	after, diagnostics := Parse(formatted)
	if diagnostics.HasErrors() || !reflect.DeepEqual(before, after.Declarations[1].(*ActivityDecl).Assembly.Spec) {
		t.Fatal("search alternatives did not round-trip", diagnostics)
	}
	fixed, err := Format(after)
	if err != nil || fixed != formatted {
		t.Fatal("search alternative formatting is not stable", err)
	}
	clone := file.Clone().Declarations[1].(*ActivityDecl)
	clone.Assembly.Spec.SearchAlternatives[0].Grammar = "changed"
	if before.SearchAlternatives[0].Grammar == "changed" {
		t.Fatal("syntax clone shares source alternative storage")
	}
	for _, pair := range [][2]string{
		{`"contextual"`, `"wider"`},
		{`grammar "integer-hole-residual/v1"`, `grammar "unknown/v1"`},
		{`max_candidates "16"`, `max_candidates "17"`},
		{`search_alternative "contextual" grammar "integer-hole-residual/v1"`, `search_alternative "contextual" grammar "integer-offset-constant/v1"`},
	} {
		_, diagnostics := Parse(strings.Replace(string(raw), pair[0], pair[1], 1))
		if !diagnostics.HasErrors() {
			t.Fatal("invalid search alternative accepted", pair)
		}
	}
}
