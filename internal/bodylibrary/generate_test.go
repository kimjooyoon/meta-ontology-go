package bodylibrary

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestLibraryDeclinesUnresolvedConstruction(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/optional-record-value-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	generated, err := Generate(context.Background(), "source.gooo", source, "Echo", "library")
	if generated != "" || err == nil || !strings.Contains(err.Error(), "authored computes bodies") {
		t.Fatal("construction was selected without an explicit evaluation route", generated, err)
	}
}

func TestLibraryNamesPreserveDistinctFieldsAndOptionalLists(t *testing.T) {
	const source = `package records
namespace records
entity Record id "urn:library:record" fields {
 field label id "urn:library:lower" type string required one
 field Label id "urn:library:upper" type string required one
 field 메모 id "urn:library:note" type string optional many
 field counts id "urn:library:counts" type integer required many
}
activity Keep(Record) -> Record id "urn:library:keep" computes "return input"
`
	generated, err := Generate(context.Background(), "source.gooo", []byte(source), "Keep", "records")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"F_label", "F_Label", "F_메모", "*[]string", "[]int64", `gooo:"urn:library:note"`, `json:"메모,omitempty"`} {
		if !strings.Contains(generated, fragment) {
			t.Fatal("native field semantics lost", fragment, generated)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if output, err := Generate(ctx, "source.gooo", []byte(source), "Keep", "records"); err == nil || output != "" {
		t.Fatal("cancelled generation produced source", output, err)
	}
}
