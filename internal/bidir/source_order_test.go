package bidir

import (
	"reflect"
	"testing"
)

func TestPutPreservesDSLInputPortSourceOrderWhenIDsDisagree(t *testing.T) {
	document := sourceOrderedInputDocument()
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckGetPut(document); err != nil {
		t.Fatal(err)
	}
	want := []ID{"billing://entity/zebra", "billing://entity/apple"}
	if got := inputIDs(written); !reflect.DeepEqual(got, want) {
		t.Fatalf("Put reordered source ports by lexical ID: got %v want %v", got, want)
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
