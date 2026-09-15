package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"testing"
)

func TestSyntaxAdapterUsesCanonicalActivityFields(t *testing.T) {
	file, diagnostics := syntax.ParseFile("aliases.gooo", `package billing
namespace billing
entity Zebra id "billing://entity/zebra"
entity Apple id "billing://entity/apple"
activity Process(Zebra, Apple) -> Zebra`)
	if diagnostics.Error() != nil {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	canonical, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	activity := file.Declarations[2].(*syntax.ActivityDecl)
	activity.Parameters = []syntax.NameRef{{Name: "WrongInput"}}
	activity.Result = syntax.NameRef{Name: "WrongOutput"}
	adapted, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(adapted, canonical) {
		t.Fatalf("legacy aliases changed canonical document:\n got %#v\nwant %#v", adapted, canonical)
	}
	got, err := Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	want, err := LowerDocument(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !EquivalentAfterRoundTrip(got, want) {
		t.Fatalf("legacy aliases changed semantic lowering")
	}
	if got := adapted.Declarations[2].Inputs; len(got) != 2 || got[0].Name != "Zebra" || got[1].Name != "Apple" {
		t.Fatalf("canonical input order was not retained: %#v", got)
	}
	if got := adapted.Declarations[2].Outputs; len(got) != 1 || got[0].Name != "Zebra" {
		t.Fatalf("canonical output identity was not retained: %#v", got)
	}
}
