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
