package duplicates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestAnalyzeFindsNormalizedDuplicateFunctions(t *testing.T) {
	root := t.TempDir()
	source := `package fixture
func first(input int) int {
	total := input + 1
	total *= 2
	return total
}
func second(input int) int {
	total := input + 1
	total *= 2
	return total
}
func distinct(input int) int {
	total := input + 2
	total *= 2
	return total
}
`
	if err := os.WriteFile(filepath.Join(root, "fixture.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	testSource := `package fixture
func testCopy(input int) int {
	total := input + 1
	total *= 2
	return total
}
`
	if err := os.WriteFile(filepath.Join(root, "fixture_test.go"), []byte(testSource), 0o600); err != nil {
		t.Fatal(err)
	}
	observations, err := Analyze(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].Dimension != sourcepolicy.DimensionRefactorDuplicate || observations[0].Value != 1 {
		t.Fatalf("unexpected duplicate observations: %#v", observations)
	}
	if !strings.Contains(observations[0].Detail, ":2:first") || !strings.Contains(observations[0].Detail, ":7:second") {
		t.Fatalf("duplicate members are not stable: %#v", observations[0])
	}
}
