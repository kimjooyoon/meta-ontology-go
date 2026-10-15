package syntaxregistration

import (
	"go/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestSelectionChunkGrowth11617(t *testing.T) {
	root := os.Getenv("GOOO_REGISTERED_SOURCE_ROOT")
	if root == "" {
		root = "../../.."
	}
	source, version, err := selectionGrowthSnapshot11617(root)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://selection-growth11617"
	if err := generateSelection(source, version+1, languagesyntax.FixedCapabilityTotal); err != nil {
		t.Fatal("owned chunks cannot accept another capability", err)
	}
	changes, err := source.finish()
	if err != nil {
		t.Fatal(err)
	}
	next, err := parseSourceUnits(changes, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	function, err := next.function("activeDenominator")
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := selectionEntries(next, function, "[]byte", "nil")
	if err != nil {
		t.Fatal("newly emitted chunk wrapper is not readable", err)
	}
	found := false
	for _, entry := range entries {
		if entry.capability == languagesyntax.FixedCapabilityTotal+1 {
			found = entry.result == evidenceName(version+1)
		}
	}
	if !found {
		t.Fatal("new capability lost its exact evidence symbol")
	}
	t.Logf("fresh selection-only continuation from v%d to v%d; exact new capability target preserved; language executions0", version, version+1)
}

func selectionGrowthSnapshot11617(root string) (*goSource, int, error) {
	paths, err := filepath.Glob(filepath.Join(root, closureRoot, "*.go"))
	if err != nil {
		return nil, 0, err
	}
	units := map[string][]byte{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, 0, err
		}
		units[closureRoot+filepath.Base(path)] = raw
	}
	source, err := parseSourceUnits(units, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		return nil, 0, err
	}
	version := 0
	for _, declaration := range source.file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range group.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range value.Names {
				if after, ok := strings.CutPrefix(name.Name, "embeddedDenominatorV"); ok {
					number, _ := strconv.Atoi(after)
					version = max(version, number)
				}
			}
		}
	}
	return source, version, nil
}
