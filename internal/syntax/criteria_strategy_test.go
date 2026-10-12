package syntax

import (
	"strings"
	"testing"
)

func TestCriteriaStrategyRoundTripAndInvalidDeclarations(t *testing.T) {
	source := strings.Replace(sourceRecordObserve, `value_case "[0]" -> "{\"value\":0}"`, `criteria { predicate "negative" activity "Negative" input "[0]" }`, 1)
	source = strings.Replace(source, `value_input "[9007199254740993]" oracle "Oracle"`, `criteria_input "[7]" criteria_strategy "distinguish_then_check"`, 1)
	file, diagnostics := ParseFileWithEntityFieldsSupport("strategy.gooo", source, EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	formatted, err := FormatWithEntityFieldsSupport(file, EntityFieldsV1Support())
	if err != nil || !strings.Contains(formatted, `criteria_strategy "distinguish_then_check"`) {
		t.Fatal(err, formatted)
	}
	other, diagnostics := ParseFileWithEntityFieldsSupport("strategy.gooo", formatted, EntityFieldsV1Support())
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	a := file.Declarations[2].(*ActivityDecl).Assembly.Spec
	b := other.Declarations[2].(*ActivityDecl).Assembly.Spec
	x, _ := a.Canonical()
	y, _ := b.Canonical()
	if x != y || a.Clone().Observation.CriteriaStrategy != "distinguish_then_check" {
		t.Fatal("strategy lost in source roundtrip")
	}
	for _, bad := range []string{`criteria_strategy ""`, `criteria_strategy "random"`, `criteria_strategy "distinguish_then_check" criteria_strategy "distinguish_then_check"`} {
		invalid := strings.Replace(source, `criteria_strategy "distinguish_then_check"`, bad, 1)
		_, diagnostics := ParseFileWithEntityFieldsSupport("strategy.gooo", invalid, EntityFieldsV1Support())
		if !diagnostics.HasErrors() {
			t.Fatal("invalid strategy accepted", bad)
		}
	}
}
