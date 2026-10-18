package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

func TestSelectionOpensOneNewChunkAtCapacity(t *testing.T) {
	entries := make([]selectionEntry, selectionChunkSize)
	for index := range entries {
		entries[index] = selectionEntry{index + 1, fmt.Sprintf("denominator%d", index+1)}
	}
	wrapper, chunks := renderSelection("activeDenominator", "[]byte", "nil", entries)
	raw, err := format.Source([]byte("package verticalsliceclosureshadow\n" + wrapper + chunks))
	if err != nil {
		t.Fatal(err)
	}
	path := closureRoot + "selection.go"
	source, err := parseSourceUnits(map[string][]byte{path: raw}, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = "gooo://syntax-registration/activity/denominator-selection"
	prior, err := source.function("activeDenominatorChunk0")
	if err != nil {
		t.Fatal(err)
	}
	priorText := source.text(prior)
	if err := extendSelection(source, "activeDenominator", "denominator17", "[]byte", "nil", 16); err != nil {
		t.Fatal(err)
	}
	changes, err := source.finish()
	if err != nil || len(changes) != 1 {
		t.Fatal("selection changed more than its source owner", err)
	}
	after, err := parseSourceUnits(changes, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	root, err := after.function("activeDenominator")
	if err != nil {
		t.Fatal(err)
	}
	next, helpers, err := selectionEntries(after, root, "[]byte", "nil")
	if err != nil || len(next) != 17 || len(helpers) != 2 || next[16] != (selectionEntry{17, "denominator17"}) {
		t.Fatal("new bounded selection chunk is not replayable", err)
	}
	if after.text(helpers[0]) != priorText {
		t.Fatal("full prior selection chunk was rewritten")
	}
}

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
	priorRoot, err := source.function("activeDenominator")
	if err != nil {
		t.Fatal(err)
	}
	priorEntries, priorHelpers, err := selectionEntries(source, priorRoot, "[]byte", "nil")
	if err != nil {
		t.Fatal(err)
	}
	priorCases := make([][]string, len(priorHelpers))
	for index, helper := range priorHelpers {
		selection := helper.Body.List[0].(*ast.SwitchStmt)
		for _, clause := range selection.Body.List {
			priorCases[index] = append(priorCases[index], source.text(clause))
		}
	}
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
	entries, helpers, err := selectionEntries(next, function, "[]byte", "nil")
	if err != nil {
		t.Fatal("newly emitted chunk wrapper is not readable", err)
	}
	if len(entries) != len(priorEntries)+1 || !reflect.DeepEqual(entries[:len(priorEntries)], priorEntries) || len(helpers) < len(priorHelpers) {
		t.Fatal("existing denominator selection entries changed")
	}
	for index, helper := range helpers[:len(priorHelpers)] {
		selection := helper.Body.List[0].(*ast.SwitchStmt)
		for row, text := range priorCases[index] {
			if next.text(selection.Body.List[row]) != text {
				t.Fatalf("prior selection chunk %d case %d was rewritten", index, row)
			}
		}
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
