package query

import (
	"testing"
)

func TestDerivedInverseRulesSeparateLayersAndStayOutOfGraphHash(t *testing.T) {
	graph := New()
	activity := id("urn:derived:activity:compile")
	entity := id("urn:derived:entity:source")
	other := id("urn:derived:activity:other")
	generated := id("urn:derived:entity:generated")
	nonIncident := id("urn:derived:activity:non-incident")
	derived := id("urn:derived:entity:derived")
	assertAdd(t, graph, NewFact(activity, Used, entity))
	assertAdd(t, graph, NewCandidateFact(other, Used, entity, "unresolved"))
	assertAdd(t, graph, NewFact(generated, WasGeneratedBy, activity))
	assertAdd(t, graph, NewFact(nonIncident, WasGeneratedBy, id("urn:derived:entity:elsewhere")))
	assertAdd(t, graph, NewFact(derived, WasDerivedFrom, entity))
	assertAdd(t, graph, NewFact(id("urn:derived:entity:unrelated"), WasDerivedFrom, id("urn:derived:entity:elsewhere")))
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	response, err := graph.Execute(derivedEnvelope(entity, RuleUsedBy, LayerAll, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 1 || len(response.Result.DerivedCandidates) != 1 {
		t.Fatalf("inverse layers were not separated: %#v", response.Result)
	}
	deterministic := response.Result.DerivedDeterministic[0]
	if deterministic.Subject != entity || deterministic.Predicate != DerivedUsedBy || deterministic.Object != activity {
		t.Fatalf("unexpected usedBy result: %#v", deterministic)
	}
	if deterministic.Status != DerivedFactStatus || deterministic.SourceLayer != FactDeterministic.String() {
		t.Fatalf("derived authority status was not explicit: %#v", deterministic)
	}
	if response.Metadata.DerivedStatus != DerivedStatusNonAuthoritative ||
		response.Metadata.DerivedRuleSchema != DerivedRuleSchemaVersion {
		t.Fatalf("derived metadata = %#v", response.Metadata)
	}
	if response.Metadata.GraphHash != beforeHash || graph.Canonical() != beforeCanonical ||
		graph.StableHash() != beforeHash {
		t.Fatal("derived evaluation mutated graph authority or hash")
	}
	label := envelopeAuthorityLabel(response.Metadata, "derived_query")
	if label.Authority != "derived" || label.Status != DerivedStatusNonAuthoritative {
		t.Fatalf("derived authority label = %#v", label)
	}

	generatedResponse, err := graph.Execute(derivedEnvelope(activity, RuleGeneratedBy, LayerDeterministic, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(generatedResponse.Result.DerivedDeterministic) != 1 ||
		generatedResponse.Result.DerivedDeterministic[0].Object != generated {
		t.Fatalf("generatedBy did not filter incident edges: %#v", generatedResponse.Result)
	}
	derivedResponse, err := graph.Execute(derivedEnvelope(entity, RuleDerivedTo, LayerDeterministic, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(derivedResponse.Result.DerivedDeterministic) != 1 ||
		derivedResponse.Result.DerivedDeterministic[0].Object != derived {
		t.Fatalf("derivedTo did not filter incident edges: %#v", derivedResponse.Result)
	}
}
