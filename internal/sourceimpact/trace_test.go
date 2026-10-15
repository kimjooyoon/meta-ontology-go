package sourceimpact

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func traceSnapshot(ids ...string) snapshot {
	s := snapshot{nodes: map[string]semantic.Node{}}
	for _, id := range ids {
		s.nodes[id] = semantic.Node{ID: semantic.ID(id), Name: id}
	}
	return s
}

func TestTraceNeverSplicesEdgesAcrossRevisions(t *testing.T) {
	before, after := traceSnapshot("a", "b", "c"), traceSnapshot("a", "b", "c")
	before.addEdge("a", "b", "body_call", "computes:0")
	after.addEdge("b", "c", "body_call", "computes:0")
	r := Report{Changes: []Change{{ID: "a"}}}
	if err := traceImpacts(context.Background(), &r, before, after); err != nil {
		t.Fatal(err)
	}
	if len(r.Affected) != 2 || r.Affected[1].ID != "b" || len(r.Affected[1].Causes) != 1 || r.Affected[1].Causes[0].Revision != "before" {
		t.Fatal("fabricated a path from different revisions", r)
	}
}

func TestTraceCyclesAndNearestChangeHaveBoundedExplanations(t *testing.T) {
	s := traceSnapshot("a", "b", "c", "d")
	s.addEdge("a", "b", "body_call", "computes:0")
	s.addEdge("b", "a", "body_call", "computes:0")
	s.addEdge("b", "c", "body_call", "computes:1")
	s.addEdge("c", "d", "body_call", "computes:0")
	r := Report{Changes: []Change{{ID: "a"}, {ID: "c"}}}
	if err := traceImpacts(context.Background(), &r, s, s); err != nil {
		t.Fatal(err)
	}
	if len(r.Affected) != 4 {
		t.Fatal(r)
	}
	for _, row := range r.Affected {
		if len(row.Causes) != 2 {
			t.Fatal(row)
		}
	}
	if r.Affected[3].Causes[0].ChangedID != "c" || r.Affected[3].Causes[0].Depth != 1 {
		t.Fatal(r.Affected[3])
	}
}

func TestRemovedAndAddedDeclarationsKeepRevisionPresence(t *testing.T) {
	before, after := traceSnapshot("removed"), traceSnapshot("added")
	r := Report{Changes: []Change{{ID: "added"}, {ID: "removed"}}}
	if err := traceImpacts(context.Background(), &r, before, after); err != nil {
		t.Fatal(err)
	}
	if len(r.Affected) != 2 || !r.Affected[0].PresentAfter || r.Affected[1].PresentAfter || r.Affected[1].Causes[0].Revision != "before" {
		t.Fatal(r)
	}
}
