package languagedeterministicquery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
)

type graphFixture struct {
	Nodes             []queryengine.Node
	Facts             []queryengine.Fact
	Candidates        []queryengine.Fact
	Targets           map[string]queryengine.ID
	ConceptID         queryengine.ID
	OperationID       queryengine.ID
	CandidateEntityID queryengine.ID
	CandidateActionID queryengine.ID
}

func buildFixture(concept languageconcept.Concept) (graphFixture, error) {
	fixture := graphFixture{Targets: make(map[string]queryengine.ID),
		ConceptID:   stableID("concept", concept.ID),
		OperationID: stableID("operation", concept.MetaOperation)}
	fixture.Nodes = append(fixture.Nodes, node(fixture.ConceptID, queryengine.EntityNodeKind, concept.ID), node(fixture.OperationID, queryengine.ActivityNodeKind, concept.MetaOperation))
	fixture.Facts = append(fixture.Facts, queryengine.NewFact(fixture.ConceptID, queryengine.WasGeneratedBy, fixture.OperationID))
	appendBindings(&fixture, BindingCode, concept.CodeBindings)
	appendBindings(&fixture, BindingMetric, concept.MetricBindings)
	useCases := make([]string, 0, len(concept.UseCases))
	for _, useCase := range concept.UseCases {
		useCases = append(useCases, useCase.ID)
	}
	appendBindings(&fixture, BindingUseCase, useCases)
	fixture.Targets[targetKey(BindingConcept, concept.ID)] = fixture.ConceptID
	appendCandidateLaw(&fixture)
	if len(fixture.Nodes) == 0 || len(fixture.Facts) == 0 {
		return graphFixture{}, fmt.Errorf("empty deterministic query fixture")
	}
	return fixture, nil
}

func appendBindings(fixture *graphFixture, class string, bindings []string) {
	for _, binding := range bindings {
		id := stableID(class, binding)
		fixture.Nodes = append(fixture.Nodes, node(id, queryengine.EntityNodeKind, binding))
		fixture.Facts = append(fixture.Facts, queryengine.NewFact(fixture.OperationID, queryengine.Used, id))
		fixture.Targets[targetKey(class, binding)] = id
	}
}

func appendCandidateLaw(fixture *graphFixture) {
	fixture.CandidateEntityID = stableID("law-candidate-entity", ConceptID)
	fixture.CandidateActionID = stableID("law-candidate-action", ConceptID)
	fixture.Nodes = append(fixture.Nodes, node(fixture.CandidateEntityID, queryengine.EntityNodeKind, "candidate-entity"), node(fixture.CandidateActionID, queryengine.ActivityNodeKind, "candidate-action"))
	fixture.Candidates = append(fixture.Candidates, queryengine.NewCandidateFact(fixture.CandidateEntityID, queryengine.WasGeneratedBy, fixture.CandidateActionID, "law-only-candidate"))
}

func stableID(kind, value string) queryengine.ID {
	sum := sha256.Sum256([]byte(kind + "\x00" + value))
	raw := "urn:gooo:" + kind + ":" + hex.EncodeToString(sum[:])
	id, _ := queryengine.NewID(raw)
	return id
}

func node(id queryengine.ID, kind queryengine.NodeKind, name string) queryengine.Node {
	return queryengine.Node{ID: id, Kind: kind, Namespace: "gooo-meta", Name: name}
}

func targetKey(class, binding string) string {
	return class + "\x00" + binding
}
