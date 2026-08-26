package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
	"strings"
	"testing"
)

func TestSyntaxAdapterRejectsLegacyOnlyAliasesWithoutMutation(t *testing.T) {
	file, diagnostics := syntax.Parse(`package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order`)
	if diagnostics.Error() != nil {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	activity := file.Declarations[1].(*syntax.ActivityDecl)
	activity.Inputs = nil
	activity.Output = ""
	before := *activity
	before.Inputs = append([]syntax.NameRef(nil), activity.Inputs...)
	before.Parameters = append([]syntax.NameRef(nil), activity.Parameters...)
	if _, err := DocumentFromSyntax(file); err == nil || !strings.Contains(err.Error(), "legacy-only Parameters") {
		t.Fatalf("legacy-only aliases were not rejected deterministically: %v", err)
	}
	if !reflect.DeepEqual(*activity, before) {
		t.Fatalf("legacy-only rejection mutated syntax state: %#v", activity)
	}
	file, diagnostics = syntax.Parse(`package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order`)
	if diagnostics.Error() != nil {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	activity = file.Declarations[1].(*syntax.ActivityDecl)
	activity.Output = ""
	before = *activity
	before.Inputs = append([]syntax.NameRef(nil), activity.Inputs...)
	before.Parameters = append([]syntax.NameRef(nil), activity.Parameters...)
	if _, err := DocumentFromSyntax(file); err == nil || !strings.Contains(err.Error(), "legacy-only Result") {
		t.Fatalf("legacy-only result was not rejected deterministically: %v", err)
	}
	if !reflect.DeepEqual(*activity, before) {
		t.Fatalf("legacy-only result rejection mutated syntax state: %#v", activity)
	}
}
func TestTypedLowererRetainsOutputPortSpansAndOrder(t *testing.T) {
	document := sourceOrderedOutputDocument()
	first, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LowerDocument(document)
	if err != nil {
		t.Fatal(err)
	}
	firstOutputs := typedOutputFacts(first)
	secondOutputs := typedOutputFacts(second)
	if !reflect.DeepEqual(firstOutputs, secondOutputs) {
		t.Fatalf("repeated typed lowering changed output evidence: %#v != %#v", firstOutputs, secondOutputs)
	}
	wantIDs := []semantic.ID{"billing://entity/zebra", "billing://entity/apple"}
	if got := outputFactIDsBySpan(firstOutputs); !reflect.DeepEqual(got, wantIDs) {
		t.Fatalf("typed lowering lost authoritative output order: got %v want %v", got, wantIDs)
	}
	if len(firstOutputs) != 2 || firstOutputs[0].Span == firstOutputs[1].Span {
		t.Fatalf("typed lowering did not retain two distinct output spans: %#v", firstOutputs)
	}
}
