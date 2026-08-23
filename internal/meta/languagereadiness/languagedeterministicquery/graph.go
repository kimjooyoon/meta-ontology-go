package languagedeterministicquery

import queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"

func instantiateGraph(fixture graphFixture, reverse bool) (*queryengine.Graph, error) {
	graph := queryengine.New()
	for _, index := range indexes(len(fixture.Nodes), reverse) {
		if err := graph.AddNode(fixture.Nodes[index]); err != nil {
			return nil, err
		}
	}
	for _, index := range indexes(len(fixture.Facts), reverse) {
		if err := graph.AddDeterministic(fixture.Facts[index]); err != nil {
			return nil, err
		}
	}
	for _, index := range indexes(len(fixture.Candidates), reverse) {
		if err := graph.AddCandidate(fixture.Candidates[index]); err != nil {
			return nil, err
		}
	}
	return graph, nil
}

func indexes(length int, reverse bool) []int {
	result := make([]int, length)
	for index := range result {
		if reverse {
			result[index] = length - index - 1
		} else {
			result[index] = index
		}
	}
	return result
}

func requestFor(definition Definition, fixture graphFixture) queryengine.Request {
	root := fixture.OperationID
	target := fixture.Targets[targetKey(definition.BindingClass, definition.Binding)]
	relation := queryengine.Used
	if definition.BindingClass == BindingConcept {
		root, target, relation = fixture.ConceptID, fixture.OperationID, queryengine.WasGeneratedBy
	}
	return exactRequest(root, target, relation, queryengine.LayerDeterministic)
}

func exactRequest(root, target queryengine.ID, relation queryengine.Relation, layer queryengine.Layer) queryengine.Request {
	return queryengine.Request{
		Schema: queryengine.QueryEnvelopeSchema, Operation: queryengine.OperationExact,
		Root: root, Target: target, Relation: relation, Layer: layer, Limit: 1,
	}
}
