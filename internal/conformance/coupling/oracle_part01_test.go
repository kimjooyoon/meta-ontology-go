package coupling

import (
	"reflect"
	"strings"
	"testing"
)

func TestCorpusExpectations(t *testing.T) {
	for _, row := range testCorpus() {
		got := Evaluate(row.Input)
		if got.Decision != row.Expected.Decision || got.Reason != row.Expected.Reason ||
			!reflect.DeepEqual(got.AcceptedSurfaces, row.Expected.AcceptedSurfaces) ||
			!reflect.DeepEqual(got.ChangedSurfaces, row.Expected.ChangedSurfaces) ||
			!reflect.DeepEqual(got.ReceiptSurfaces, row.Expected.ReceiptSurfaces) ||
			got.ObservationCounts != row.Expected.ObservationCounts || got.Resources != row.Expected.Resources {
			t.Errorf("%s got decision=%s/%s accepted=%v/%v changed=%v/%v receipts=%v/%v counts=%+v/%+v resources=%+v/%+v", row.Name, got.Decision, got.Reason, got.AcceptedSurfaces, row.Expected.AcceptedSurfaces, got.ChangedSurfaces, row.Expected.ChangedSurfaces, got.ReceiptSurfaces, row.Expected.ReceiptSurfaces, got.ObservationCounts, row.Expected.ObservationCounts, got.Resources, row.Expected.Resources)
		}
	}
}
func TestBaselineIsFullSuiteAndFair(t *testing.T) {
	for _, row := range testCorpus()[:4] {
		comparison := Compare(row.Input)
		if !comparison.OutcomeMatch || !comparison.ReasonMatch || !comparison.LocalizationMatch || comparison.Finding != "NO_UNIQUE_BENEFIT" {
			t.Fatalf("%s comparison=%+v", row.Name, comparison)
		}
		if !comparison.Baseline.FullSuite || comparison.Baseline.Resources != comparison.Oracle.Resources || comparison.Baseline.WorkUnits != comparison.Oracle.Resources.WorkUnits {
			t.Fatalf("%s baseline resource binding=%+v oracle=%+v", row.Name, comparison.Baseline, comparison.Oracle.Resources)
		}
	}
}
func TestStrictJSONAndPermutationInvariance(t *testing.T) {
	input := testCorpus()[0].Input
	data, err := EncodeInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInput(data)
	if err != nil {
		t.Fatal(err)
	}
	if CanonicalInputDigest(input) != CanonicalInputDigest(decoded) || Evaluate(input).CanonicalOutputDigest != Evaluate(decoded).CanonicalOutputDigest {
		t.Fatal("JSON round trip changed canonical result")
	}
	if _, err := DecodeInput(append(data, []byte("\n{}")...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema": "gooo/coupling/v1"`, `"schema": "gooo/coupling/v1","schema": "gooo/coupling/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "}") + `,"unexpected":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}

	permuted := cloneInput(input)
	reverseBindings(permuted.Registry)
	reverseChanges(permuted.Changes)
	reverseReceipts(permuted.Receipts)
	reverseResources(permuted.ResourceReceipts)
	reverseStrings(permuted.Roots)
	reverseEdges(permuted.Path.Edges)
	reverseClaims(permuted.Path.Claims)
	reverseEvidence(permuted.Path.Evidence)
	if CanonicalInputDigest(input) != CanonicalInputDigest(permuted) || !reflect.DeepEqual(Evaluate(input), Evaluate(permuted)) {
		t.Fatal("presentation ordering changed canonical result")
	}
}
