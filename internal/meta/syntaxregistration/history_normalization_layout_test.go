package syntaxregistration

import (
	"bytes"
	"fmt"
	"go/ast"
	"reflect"
	"testing"
)

func historyNormalizationLayout(t *testing.T, units int) *goSource {
	t.Helper()
	initial := parseHistoryCapacityFixture(t, historyCapacityLegacyFixture())
	history, err := readHistoryOrigins(initial)
	if err != nil {
		t.Fatal(err)
	}
	writeHistoryOrigins(initial, history)
	changes, err := initial.finish()
	if err != nil {
		t.Fatal(err)
	}
	partitioned := parseHistoryCapacityFixture(t, changes[closureRoot+"history.go"])
	inputs := map[string][]byte{}
	for index, declaration := range partitioned.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			t.Fatal("fixture declaration is not a function")
		}
		path := fmt.Sprintf("%sownership_unit_%d.go", closureRoot, index%units)
		if inputs[path] == nil {
			inputs[path] = []byte("package verticalsliceclosureshadow\n")
		}
		inputs[path] = append(inputs[path], []byte(partitioned.text(function)+"\n")...)
	}
	inputs[closureRoot+"untouched.go"] = []byte("package verticalsliceclosureshadow\nconst ownershipSentinel = 20261015\n")
	source, err := parseSourceUnits(inputs, closureRoot, "verticalsliceclosureshadow", false)
	if err != nil {
		t.Fatal(err)
	}
	source.activity = initial.activity
	return source
}

func TestHistoryNormalizationFollowsDeclarationOwners(t *testing.T) {
	for _, count := range []int{1, 2, 5} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			source := historyNormalizationLayout(t, count)
			inputs := map[string][]byte{}
			for path, raw := range source.units {
				inputs[path] = bytes.Clone(raw)
			}
			prior, err := readHistoryOrigins(source)
			if err != nil {
				t.Fatal(err)
			}
			writeHistoryOrigins(source, prior)
			changes, err := source.finish()
			if err != nil || len(changes) != count {
				t.Fatal("fixture ownership did not span its selected units", len(changes), err)
			}
			updated := historyNormalizationUnits(t, source, prior, changes)
			after, err := parseSourceUnits(updated, closureRoot, "verticalsliceclosureshadow", false)
			if err != nil {
				t.Fatal(err)
			}
			observed, err := readHistoryOrigins(after)
			if err != nil || !reflect.DeepEqual(prior.targets, observed.targets) {
				t.Fatal("physical layout changed logical history", err)
			}
			if !reflect.DeepEqual(inputs, source.units) ||
				string(updated[closureRoot+"untouched.go"]) != string(inputs[closureRoot+"untouched.go"]) {
				t.Fatal("normalization changed input or an unrelated source unit")
			}
			t.Logf("%d owned units; %d targets retained; unrelated unit unchanged; native/model executions 0", count, len(observed.targets))
		})
	}
}
