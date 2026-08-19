package query

import (
	"testing"
)

func TestDatalogPositiveRulesReachFixedPointInStableOrder(t *testing.T) {
	facts := []Fact{
		NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")),
		NewFact(id("urn:entity:b"), WasDerivedFrom, id("urn:entity:c")),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	rules := []Rule{
		{
			ID:   "transitive/v1",
			Head: Triple("dependsOn", Variable("subject"), Variable("source")),
			Body: []Atom{
				Triple("wasDerivedFrom", Variable("subject"), Variable("middle")),
				Triple("dependsOn", Variable("middle"), Variable("source")),
			},
		},
		{
			ID:   "direct/v1",
			Head: Triple("dependsOn", Variable("subject"), Variable("source")),
			Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		},
	}
	request := DatalogQuery{
		Patterns:       []Atom{Triple("dependsOn", Variable("subject"), Variable("source"))},
		Rules:          rules,
		IncludeDerived: true,
	}
	left, err := first.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	right, err := second.QueryDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(left.Derived) != 3 || len(left.Rows) != 3 || !left.Complete {
		t.Fatalf("fixed point result = %#v", left)
	}
	want := [][2]ID{
		{id("urn:entity:a"), id("urn:entity:b")},
		{id("urn:entity:a"), id("urn:entity:c")},
		{id("urn:entity:b"), id("urn:entity:c")},
	}
	for index, row := range left.Rows {
		subject, subjectOK := row.Value("?subject")
		source, sourceOK := row.Value("source")
		if !subjectOK || !sourceOK || subject != want[index][0] || source != want[index][1] {
			t.Fatalf("row %d = %#v, want %v", index, row, want[index])
		}
	}
	if left.Canonical() != right.Canonical() || left.StableHash() != right.StableHash() {
		t.Fatalf("permuted Datalog replay changed result: %s/%s vs %s/%s", left.Canonical(), left.StableHash(), right.Canonical(), right.StableHash())
	}
	if left.Derived[1].RuleID != "transitive/v1" || len(left.Derived[1].Support) != 2 {
		t.Fatalf("derived proof was not retained: %#v", left.Derived[1])
	}
}
