package query

import (
	"errors"
	"reflect"
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

func TestDatalogConjunctionPermutationHasOneCanonicalResult(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:activity:start"), Used, id("urn:entity:input")))
	assertAdd(t, graph, NewFact(id("urn:activity:start"), WasAssociatedWith, id("urn:agent:operator")))
	rule := Rule{
		ID:   "annotated/v1",
		Head: Triple("annotated", Variable("activity"), Variable("agent")),
		Body: []Atom{
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
			Triple("used", Variable("activity"), Variable("input")),
		},
	}
	forward, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{
			Triple("used", Variable("activity"), Variable("input")),
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
		},
		Rules: []Rule{rule}, IncludeDerived: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
			Triple("used", Variable("activity"), Variable("input")),
		},
		Rules: []Rule{{
			ID:   rule.ID,
			Head: rule.Head,
			Body: []Atom{rule.Body[1], rule.Body[0]},
		}}, IncludeDerived: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if forward.Canonical() != reversed.Canonical() || forward.StableHash() != reversed.StableHash() {
		t.Fatalf(
			"conjunction permutation changed canonical result: %s/%s vs %s/%s",
			forward.Canonical(), forward.StableHash(), reversed.Canonical(), reversed.StableHash(),
		)
	}
	if len(forward.Derived) != 1 || len(forward.Derived[0].Support) != 2 ||
		forward.Derived[0].Support[0].Predicate != string(Used) {
		t.Fatalf("derived support was not normalized: %#v", forward.Derived)
	}
}

func TestDatalogCandidatesAreVisibleButNeverRuleInputs(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")))
	assertAdd(t, graph, NewCandidateFact(id("urn:entity:b"), WasDerivedFrom, id("urn:entity:c"), "unresolved"))
	rule := Rule{
		ID:   "direct/v1",
		Head: Triple("dependsOn", Variable("subject"), Variable("source")),
		Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
	}
	derived, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("dependsOn", Variable("subject"), Variable("source"))},
		Rules:    []Rule{rule}, IncludeDerived: true, IncludeCandidates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(derived.Derived) != 1 || len(derived.Rows) != 1 {
		t.Fatalf("candidate fact entered rule closure: %#v", derived)
	}
	if derived.Derived[0].Origin != DatalogDerived || derived.Derived[0].Subject != id("urn:entity:a") {
		t.Fatalf("derived origin = %#v", derived.Derived[0])
	}
	candidates, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns:          []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		IncludeCandidates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates.Rows) != 2 || candidates.Rows[0].Facts[0].Origin != DatalogDeclared ||
		candidates.Rows[1].Facts[0].Origin != DatalogCandidate {
		t.Fatalf("candidate visibility/order = %#v", candidates)
	}
}

func TestDatalogRejectsUnsafeOrUnknownQueriesBeforeEvaluation(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")))
	base := DatalogQuery{Patterns: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))}}
	cases := []struct {
		name string
		edit func(*DatalogQuery)
	}{
		{"no patterns", func(query *DatalogQuery) { query.Patterns = nil }},
		{"unknown query predicate", func(query *DatalogQuery) {
			query.Patterns = []Atom{Triple("notDeclared", Variable("subject"), Variable("source"))}
		}},
		{"invalid variable", func(query *DatalogQuery) {
			query.Patterns = []Atom{Triple("wasDerivedFrom", Variable("?"), Variable("source"))}
		}},
		{"invalid constant", func(query *DatalogQuery) {
			query.Patterns = []Atom{Triple("wasDerivedFrom", Constant(ID("display-name")), Variable("source"))}
		}},
		{"unsafe head", func(query *DatalogQuery) {
			query.Rules = []Rule{{ID: "unsafe", Head: Triple("dependsOn", Variable("missing"), Variable("source")), Body: []Atom{
				Triple("wasDerivedFrom", Variable("subject"), Variable("source")),
			}}}
		}},
		{"unknown body predicate", func(query *DatalogQuery) {
			query.Rules = []Rule{{ID: "unknown", Head: Triple("dependsOn", Variable("subject"), Variable("source")), Body: []Atom{
				Triple("notDeclared", Variable("subject"), Variable("source")),
			}}}
			query.Patterns = []Atom{Triple("dependsOn", Variable("subject"), Variable("source"))}
		}},
		{"duplicate rule ID", func(query *DatalogQuery) {
			query.Rules = []Rule{
				{ID: "same", Head: Triple("dependsOn", Variable("subject"), Variable("source")), Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))}},
				{ID: "same", Head: Triple("dependsOn", Variable("subject"), Variable("source")), Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))}},
			}
		}},
	}
	for _, testCase := range cases {
		request := base
		testCase.edit(&request)
		if _, err := graph.EvaluateDatalog(request); !errors.Is(err, ErrInvalidDatalogQuery) {
			t.Errorf("%s error = %v, want ErrInvalidDatalogQuery", testCase.name, err)
		}
	}
}

func TestDatalogResultLimitIsDeterministicAndMarkedIncomplete(t *testing.T) {
	graph := New()
	for _, target := range []ID{id("urn:entity:z"), id("urn:entity:a"), id("urn:entity:m")} {
		assertAdd(t, graph, NewFact(id("urn:activity:start"), Used, target))
	}
	before := graph.Canonical()
	result, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("used", Constant(id("urn:activity:start")), Variable("entity"))},
		Limit:    2,
	})
	if !errors.Is(err, ErrDatalogBudget) {
		t.Fatalf("row limit error = %v, want ErrDatalogBudget", err)
	}
	if result.Complete || len(result.Rows) != 2 {
		t.Fatalf("bounded result = %#v", result)
	}
	if got, _ := result.Rows[0].Value("entity"); got != id("urn:entity:a") {
		t.Fatalf("first bounded row = %s", got)
	}
	if graph.Canonical() != before {
		t.Fatal("Datalog evaluation mutated graph authority")
	}
	if !reflect.DeepEqual(result.Rows[0].Bindings, map[string]ID{"entity": id("urn:entity:a")}) {
		t.Fatalf("binding escaped canonical shape: %#v", result.Rows[0].Bindings)
	}
}

func TestDatalogExcludedDerivedViewDoesNotConsumeClosureBudget(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")))
	assertAdd(t, graph, NewFact(id("urn:entity:b"), WasDerivedFrom, id("urn:entity:c")))
	result, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		Rules: []Rule{{
			ID:   "direct/v1",
			Head: Triple("dependsOn", Variable("subject"), Variable("source")),
			Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		}},
		MaxDerivedFacts: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 || len(result.Derived) != 0 || !result.Complete {
		t.Fatalf("excluded derived view was evaluated: %#v", result)
	}
}
