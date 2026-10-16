package main

import (
	"encoding/json"
	"testing"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func TestObservedAdmissionMutationGuardsFresh1898(t *testing.T) {
	f, baseline := admissionFixture1894(t)
	if _, err := observedAdmissionVariant(variant{}, f, countercase{}); err == nil {
		t.Fatal("absent baseline accepted")
	}
	if _, err := observedAdmissionVariant(baseline, f, countercase{Kind: "UNDECLARED"}); err == nil {
		t.Fatal("undeclared mutation kind accepted")
	}
	if _, err := admissionDocument(baseline.Inputs.Observation.Document, "UNDECLARED"); err == nil {
		t.Fatal("undeclared document mutation accepted")
	}
	if _, err := admissionReceipt(baseline.Receipt, "UNDECLARED"); err == nil {
		t.Fatal("undeclared receipt mutation accepted")
	}
	for _, raw := range []string{"{", `{}`, `{"decision":{"value":{"readiness_credit":true}}}`} {
		if _, err := admissionDocument([]byte(raw), "wrong-head"); err == nil {
			t.Fatal("document without source-bound row accepted")
		}
		if _, err := admissionReceipt([]byte(raw), "resealed-source-value"); err == nil {
			t.Fatal("receipt without exactly one false credit claim accepted")
		}
	}
}

func TestObservedAdmissionDocumentShapeFresh1899(t *testing.T) {
	_, baseline := admissionFixture1894(t)
	for _, shape := range []string{"short-head", "empty-rows", "missing-binding"} {
		t.Run(shape, func(t *testing.T) {
			var document co.ObservationDocument
			if err := json.Unmarshal(baseline.Inputs.Observation.Document, &document); err != nil {
				t.Fatal(err)
			}
			switch shape {
			case "short-head":
				document.SubjectSHA = "a"
			case "empty-rows":
				document.Observations = nil
			case "missing-binding":
				document.Observations[0].Binding = nil
			}
			if _, err := admissionDocument(marshalAdmission1894(t, document), "wrong-head"); err == nil {
				t.Fatal("malformed baseline was converted into a control")
			}
		})
	}
}

func TestObservedAdmissionCreditObservationFresh1900(t *testing.T) {
	for _, row := range []struct {
		id, status string
		want       bool
	}{
		{"META-CONCEPT-GOVERNED-REFACTORING", "SATISFIED", true},
		{"META-CONCEPT-GOVERNED-REFACTORING", "UNSATISFIED", false},
		{"UNRELATED", "SATISFIED", false},
	} {
		var receipt readiness.Receipt
		raw := `{"snapshot":{"obligations":[{"id":"` + row.id + `","status":"` + row.status + `"}]}}`
		if err := json.Unmarshal([]byte(raw), &receipt); err != nil || admissionCredit(receipt) != row.want {
			t.Fatal("readiness credit observation changed", err)
		}
	}
	if admissionCredit(readiness.Receipt{}) {
		t.Fatal("zero receipt received credit")
	}
}
