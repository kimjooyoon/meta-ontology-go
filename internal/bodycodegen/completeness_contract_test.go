package bodycodegen

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func TestGeneratedAndFailedBodiesUseDeclaredSharedReceipt(t *testing.T) {
	source := []byte("package shared\nnamespace shared\nentity Integer id \"gooo://shared/integer\"\nactivity Identity(Integer) -> Integer computes \"return input\"\n")
	result, err := Generate("shared.gooo", source, "Identity")
	if err != nil {
		t.Fatal(err)
	}
	for _, receipt := range []*CompletenessReceipt{result.Report.CompletenessReceipt, FailureCompletenessReceipt("Identity", source, "independent parse failure")} {
		if err := completeness.Validate(receipt); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(receipt)
		if err != nil {
			t.Fatal(err)
		}
		observed, err := completeness.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if *observed.FirstUnresolved != *receipt.FirstUnresolved {
			t.Fatal("consumer changed first unresolved frontier")
		}
		if completenessDimensionByID(t, receipt, "receipt_schema_binding").Status != "PASS" {
			t.Fatal("declaration/generated schema is unbound")
		}
	}
}
