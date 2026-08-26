package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"sort"
	"testing"
)

func TestApplyFixPlanIREvidenceRefsAreInsertionIndependent(t *testing.T) {
	firstIR := lowerInspectFixtureIR(t)
	secondIR := lowerInspectFixtureIR(t)
	fact := firstIR.Graph.DeterministicFacts()[0]
	evidenceA := mustAnalyzeEvidence(t, "billing://evidence/a", fact.Key())
	evidenceB := mustAnalyzeEvidence(t, "billing://evidence/b", fact.Key())
	if err := firstIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	if err := firstIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceA); err != nil {
		t.Fatal(err)
	}
	if err := secondIR.AddEvidence(evidenceB); err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile("fixture.gooo", sourceOrderA)
	if diagnostics.Error() != nil {
		t.Fatal(diagnostics.Error())
	}
	firstPlan := newFixPlan([]byte(sourceOrderA), nil, file)
	secondPlan := newFixPlan([]byte(sourceOrderA), nil, file)
	applyFixPlanIR(&firstPlan, firstIR)
	applyFixPlanIR(&secondPlan, secondIR)
	if !reflect.DeepEqual(firstPlan.Evidence, secondPlan.Evidence) || !sort.StringsAreSorted(firstPlan.Evidence.Refs) {
		t.Fatalf("evidence refs depend on insertion order: first=%#v second=%#v", firstPlan.Evidence, secondPlan.Evidence)
	}
}
