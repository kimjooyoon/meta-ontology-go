package resourcevector

import (
	"reflect"
	"strings"
	"testing"
)

func TestStrictJSONAndExpectedIsolation(t *testing.T) {
	input := R4F01()
	data, err := EncodeInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := DecodeInput(data)
	if err != nil || CanonicalInputDigest(roundTrip) != CanonicalInputDigest(input) {
		t.Fatalf("strict input round trip changed digest: err=%v", err)
	}
	if _, err := DecodeInput(append(data, []byte(` {"schema":"extra"}`)...)); err == nil {
		t.Fatal("trailing JSON value accepted")
	}
	duplicate := strings.Replace(string(data), `"schema": "gooo/selective-ci-resource-vector/v1"`, `"schema": "gooo/selective-ci-resource-vector/v1","schema": "gooo/selective-ci-resource-vector/v1"`, 1)
	if _, err := DecodeInput([]byte(duplicate)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	unknown := strings.TrimSuffix(strings.TrimSpace(string(data)), "}") + `,"unknown":true}`
	if _, err := DecodeInput([]byte(unknown)); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
	left := Evaluate(input)
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	corpus.Cases[0].Expected.Decision = DecisionFailClosed
	right := Evaluate(corpus.Cases[0].Input)
	if !reflect.DeepEqual(left, right) {
		t.Fatal("expected-only mutation changed replay output")
	}
	outputJSON, err := EncodeOutputJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(outputJSON), "promotion_authorized") || left.PromotionAuthorized() {
		t.Fatal("promotion authorization escaped the oracle")
	}
}
