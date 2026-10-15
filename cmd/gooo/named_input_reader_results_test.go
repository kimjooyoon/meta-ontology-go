package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func assertNamedReaderResults(t *testing.T, raw []byte) {
	t.Helper()
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil || receipt.Result == nil || receipt.Decision != "OBSERVED" {
		t.Fatal("native input-only reader receipt", err)
	}
	r := receipt.Result.Runtime
	if r.FiniteTotal != 0 || r.ModelCalls != 0 || len(r.Traces) != 2 {
		t.Fatal("input-only observation scope", r.FiniteTotal, r.ModelCalls, len(r.Traces))
	}
	labels := []string{"원금: Integer [input]", "원금: Integer [input0]\n공제: Integer [input1]"}
	ids := []string{"urn:gooo:example:named-inputs:credit", "urn:gooo:example:named-inputs:net"}
	source, err := os.ReadFile("../../examples/named-inputs/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(source))
	for i, trace := range r.Traces {
		var actual struct {
			Status, Reason, Labels string
			ActivityID             string `json:"activity_id"`
			SourceSHA256           string `json:"source_sha256"`
			Total, Named           int
			NamedBPS               int `json:"named_bps"`
		}
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil {
			t.Fatal("reader delivery", i)
		}
		if actual.Status != "ALL_INPUTS_NAMED" || actual.Reason != "" || actual.Total != i+1 ||
			actual.Named != i+1 || actual.NamedBPS != 10000 || actual.Labels != labels[i] ||
			actual.ActivityID != ids[i] || actual.SourceSHA256 != digest {
			t.Fatal("source-owned input description", i, actual)
		}
	}
	t.Log("two new compiler-exported signatures, two external output comparisons; finite denominator 0, model calls 0")
}
