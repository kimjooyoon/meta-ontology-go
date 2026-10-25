package syntaxregistration

import (
	"strings"
	"testing"
	"testing/fstest"
)

func partitionNativeSource(t *testing.T, files fstest.MapFS) *goSource {
	t.Helper()
	inputs := map[string][]byte{}
	for name, file := range files {
		inputs[name] = file.Data
	}
	source, err := parseSourceUnits(inputs, syntaxRoot, "languagesyntax", false)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

// Resolve the field's actual literal origin even when the projector has moved
// the declaration into a helper. Incidental occurrences are not mutation sites.
func partitionNativeFieldOwner(t *testing.T, files fstest.MapFS, field, element, needle string) string {
	t.Helper()
	source := partitionNativeSource(t, files)
	root, err := source.function("expectedRegistry")
	if err != nil {
		t.Fatal(err)
	}
	leaves, err := source.registryFieldLeaves(root, field, element, false)
	if err != nil {
		t.Fatal(err)
	}
	owner := ""
	for _, leaf := range leaves {
		if !strings.Contains(source.text(leaf), needle) {
			continue
		}
		if owner != "" {
			t.Fatalf("ambiguous %s counterexample origin", field)
		}
		owner, _ = source.location(leaf.Pos())
	}
	if owner == "" {
		t.Fatalf("missing %s counterexample origin", field)
	}
	return owner
}

func partitionNativeCounterexampleOwner(t *testing.T, files fstest.MapFS, kind, needle string) string {
	t.Helper()
	switch kind {
	case "family":
		return partitionNativeFieldOwner(t, files, "CaseSources", "CaseSourceSet", needle)
	case "meta":
		return partitionNativeFieldOwner(t, files, "MetaSources", "string", needle)
	case "projection":
		return partitionNativeFieldOwner(t, files, "ProjectionUnits", "ProjectionDefinition", needle)
	}
	source := partitionNativeSource(t, files)
	function, err := source.function("expectedSyntaxPackageUnits")
	if err != nil {
		t.Fatal(err)
	}
	name, _ := source.location(function.Pos())
	return name
}
