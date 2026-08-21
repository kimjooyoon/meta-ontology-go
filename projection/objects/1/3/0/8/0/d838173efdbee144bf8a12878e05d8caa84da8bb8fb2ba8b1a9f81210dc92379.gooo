package query

import (
	"errors"
	"testing"
)

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
