package bidir

import (
	"reflect"
	"testing"
)

func TestRelationOrderHashUsesSourceSpansAcrossPermutations(t *testing.T) {
	model := sourceOrderedRelationsModel()
	permuted := model.Clone()
	reverseRelations(permuted.Relations)
	_, leftRelations := orderedSequences(model)
	_, rightRelations := orderedSequences(permuted)
	if !reflect.DeepEqual(leftRelations, rightRelations) || sequenceHash(leftRelations) != sequenceHash(rightRelations) {
		t.Fatalf("relation order was not source-authoritative: %v != %v", leftRelations, rightRelations)
	}
}
func sourceOrderedInputDocument() Document {
	return Document{
		Package: "billing", Namespace: "billing",
		Declarations: []Declaration{
			{Kind: EntityKind, ID: "billing://entity/zebra", Name: "Zebra"},
			{Kind: EntityKind, ID: "billing://entity/apple", Name: "Apple"},
			{
				Kind: ActivityKind, ID: "billing://activity/process", Name: "Process",
				Inputs: []Reference{
					{ID: "billing://entity/zebra", Name: "Zebra", Span: SourceSpan{File: "ports.gooo", Start: 10, End: 15}},
					{ID: "billing://entity/apple", Name: "Apple", Span: SourceSpan{File: "ports.gooo", Start: 20, End: 25}},
				},
			},
		},
	}
}
func sourceOrderedOutputDocument() Document {
	document := sourceOrderedInputDocument()
	document.Declarations[2].Inputs = nil
	document.Declarations[2].Outputs = []Reference{
		{ID: "billing://entity/zebra", Name: "Zebra", Span: SourceSpan{File: "ports.gooo", Start: 30, End: 35}},
		{ID: "billing://entity/apple", Name: "Apple", Span: SourceSpan{File: "ports.gooo", Start: 40, End: 45}},
	}
	return document
}
func sourceOrderedRelationsModel() Model {
	return Model{Relations: []Relation{
		{ID: StableRelationID(PredicateInvokes, "billing://activity/zulu", "billing://activity/alpha"), Kind: PredicateInvokes, Source: "billing://activity/zulu", Target: "billing://activity/alpha", Span: SourceSpan{File: "relations.gooo", Start: 10, End: 15}},
		{ID: StableRelationID(PredicateInvokes, "billing://activity/alpha", "billing://activity/zulu"), Kind: PredicateInvokes, Source: "billing://activity/alpha", Target: "billing://activity/zulu", Span: SourceSpan{File: "relations.gooo", Start: 20, End: 25}},
	}}
}
func reverseRelations(relations []Relation) {
	for left, right := 0, len(relations)-1; left < right; left, right = left+1, right-1 {
		relations[left], relations[right] = relations[right], relations[left]
	}
}
func inputIDs(document Document) []ID {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		ids := make([]ID, len(declaration.Inputs))
		for index, input := range declaration.Inputs {
			ids[index] = input.ID
		}
		return ids
	}
	return nil
}
