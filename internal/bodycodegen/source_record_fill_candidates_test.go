package bodycodegen

import (
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestRecordFillLiteralUsesDeclaredScalarType(t *testing.T) {
	for _, test := range []struct {
		name   string
		raw    string
		typeID string
		want   string
		ok     bool
	}{
		{name: "string", raw: `"ready"`, typeID: string(semantic.BuiltinStringTypeID), want: `"ready"`, ok: true},
		{name: "integer", raw: `42`, typeID: string(semantic.BuiltinIntegerTypeID), want: `42`, ok: true},
		{name: "boolean", raw: `true`, typeID: string(semantic.BuiltinBooleanTypeID), want: `true`, ok: true},
		{name: "decimal is not integer", raw: `4.2`, typeID: string(semantic.BuiltinIntegerTypeID)},
		{name: "wrong scalar type", raw: `42`, typeID: string(semantic.BuiltinStringTypeID)},
		{name: "null", raw: `null`, typeID: string(semantic.BuiltinBooleanTypeID)},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := recordFillLiteral([]byte(test.raw), test.typeID)
			if got != test.want || ok != test.ok {
				t.Fatalf("record literal = %q, %v; want %q, %v", got, ok, test.want, test.ok)
			}
		})
	}
}

func TestRecordIntegerOrderPredicateOperatorsAreVersionedAndTyped(t *testing.T) {
	var ordered []string
	appendRecordPredicates(&ordered, map[string]bool{}, "input.score", "7", true)
	wantOrdered := []string{
		"input.score < 7", "input.score <= 7", "input.score > 7", "input.score >= 7",
		"input.score == 7", "input.score != 7",
	}
	if !slices.Equal(ordered, wantOrdered) {
		t.Fatalf("ordered record predicates = %#v, want %#v", ordered, wantOrdered)
	}

	var categorical []string
	appendRecordPredicates(&categorical, map[string]bool{}, "input.state", `"ready"`, false)
	if !slices.Equal(categorical, []string{`input.state == "ready"`, `input.state != "ready"`}) {
		t.Fatalf("categorical predicates gained ordering operators: %#v", categorical)
	}

	for _, expression := range wantOrdered {
		if selector := recordPredicateSelector(expression); selector != "input.score" {
			t.Errorf("predicate selector for %q = %q, want input.score", expression, selector)
		}
	}
}

func TestRecordIntegerRangeCompositionsPrioritizeObservedIntervals(t *testing.T) {
	var atoms []string
	seen := map[string]bool{}
	for _, value := range []string{"-1", "0", "1", "2"} {
		appendRecordPredicates(&atoms, seen, "input.score", value, true)
	}
	got := recordIntegerRangeExpressions(atoms, 16)
	want := "(input.score >= 0) && (input.score <= 1)"
	if len(got) != 16 || !slices.Contains(got, want) {
		t.Fatalf("bounded integer interval candidates omitted %q: %#v", want, got)
	}
}
