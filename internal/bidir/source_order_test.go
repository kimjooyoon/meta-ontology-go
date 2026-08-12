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
	wantSpans := []SourceSpan{{File: "ports.gooo", Start: 10, End: 15}, {File: "ports.gooo", Start: 20, End: 25}}
	if got := inputSpans(written); !reflect.DeepEqual(got, wantSpans) {
		t.Fatalf("Put dropped input evidence spans: got %#v want %#v", got, wantSpans)
	}
}

func TestPutPreservesDSLOutputPortSourceOrderWhenIDsDisagree(t *testing.T) {
	document := sourceOrderedOutputDocument()
	model, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	written, err := Put(document, model)
	if err != nil {
		t.Fatal(err)
	}
	want := []ID{"billing://entity/zebra", "billing://entity/apple"}
	if got := outputIDs(written); !reflect.DeepEqual(got, want) {
		t.Fatalf("Put reordered source ports by lexical ID: got %v want %v", got, want)
	}
	wantSpans := []SourceSpan{{File: "ports.gooo", Start: 30, End: 35}, {File: "ports.gooo", Start: 40, End: 45}}
	if got := outputSpans(written); !reflect.DeepEqual(got, wantSpans) {
		t.Fatalf("Put dropped output evidence spans: got %#v want %#v", got, wantSpans)
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

func inputSpans(document Document) []SourceSpan {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		spans := make([]SourceSpan, len(declaration.Inputs))
		for index, input := range declaration.Inputs {
			spans[index] = input.Span
		}
		return spans
	}
	return nil
}

func outputIDs(document Document) []ID {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		ids := make([]ID, len(declaration.Outputs))
		for index, output := range declaration.Outputs {
			ids[index] = output.ID
		}
		return ids
	}
	return nil
}

func outputSpans(document Document) []SourceSpan {
	for _, declaration := range document.Declarations {
		if declaration.ID != "billing://activity/process" {
			continue
		}
		spans := make([]SourceSpan, len(declaration.Outputs))
		for index, output := range declaration.Outputs {
			spans[index] = output.Span
		}
		return spans
	}
	return nil
}
