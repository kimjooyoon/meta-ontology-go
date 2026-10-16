package cliobservations

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBoundedTextPrefixPublicNativeFresh1905(t *testing.T) {
	source, err := os.ReadFile("../../examples/bounded-text/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	values := []string{"", strings.Repeat("x", 1024), strings.Repeat("한", 341) + "x", strings.Repeat("\x00", 1024)}
	for _, value := range values {
		input, _ := json.Marshal(value)
		expected, _ := json.Marshal(map[string]string{"value": "prefix" + value})
		suite.Cases = append(suite.Cases, bodyexecution.CompositionCase{
			Inputs: map[string]json.RawMessage{"app/bound1905:Prefix": input},
			Expected: map[string]json.RawMessage{"app/bound1905:Prefix": expected},
		})
	}
	receipt, code := executeTextBound1905(t, "prefix", source, "Prefix", suite, false)
	assertTextBoundNative1905(t, receipt, code, len(values))
	for i, trace := range receipt.Result.Runtime.Traces {
		if len(trace.Deliveries) != 1 || len(trace.Deliveries[0].ActualFields) != 1 {
			t.Fatal("field observations missing", i, trace)
		}
		field := trace.Deliveries[0].ActualFields[0]
		var actual string
		if field.ID != "gooo://examples/bounded-text/expanded/value" || json.Unmarshal(field.Value, &actual) != nil || actual != "prefix"+values[i] {
			t.Fatal("accepted value or field identity changed", i, field)
		}
	}
}

func TestBoundedMappedOutput13PublicNativeFresh1905(t *testing.T) {
	original, err := os.ReadFile("../bodyexecution/testdata/text-bound1905/mapped-original.gooo.fixture")
	digest := sha256.Sum256(original)
	if err != nil || hex.EncodeToString(digest[:]) != "1c23bfc07dfb68b5d03b762a572f790dcb43f2546b312dd8d02ae3ab909bdab5" {
		t.Fatal("original issue1620 source changed", err)
	}
	marker := []byte(`field canonical id "gooo://capability-probe/netip-mapped-output/result/v1/canonical" type string required one`)
	if bytes.Count(original, marker) != 1 {
		t.Fatal("original canonical field declaration is ambiguous")
	}
	source := bytes.Replace(original, marker, append(append([]byte(nil), marker...), []byte(` max_utf8_bytes "1030"`)...), 1)
	zone := strings.Repeat("z", 1007)
	input, _ := json.Marshal("::ffff:ffff:ffff%" + zone)
	expected := map[string]any{"parsed": true, "canonical": "::ffff:255.255.255.255%" + zone,
		"family": 6, "zone": zone, "is4_in6": true, "group0": 0, "group1": 0, "group2": 0,
		"group3": 0, "group4": 0, "group5": 65535, "group6": 65535, "group7": 65535}
	raw, _ := json.Marshal(expected)
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs: map[string]json.RawMessage{"app/bound1905:MappedOutput": input},
		Expected: map[string]json.RawMessage{"app/bound1905:MappedOutput": raw},
	}}}
	receipt, code := executeTextBound1905(t, "mapped13", source, "MappedOutput", suite, false)
	assertTextBoundNative1905(t, receipt, code, 1)
	delivery := receipt.Result.Runtime.Traces[0].Deliveries[0]
	if len(delivery.ActualFields) != 13 {
		t.Fatal("original field population changed", delivery.ActualFields)
	}
	seen := map[string]bool{}
	for _, field := range delivery.ActualFields {
		value, declared := expected[field.Name]
		want, _ := json.Marshal(value)
		if !declared || seen[field.Name] || field.ID != "gooo://capability-probe/netip-mapped-output/result/v1/"+field.Name || !bytes.Equal(field.Value, want) {
			t.Fatal("an original field identity or typed value differs", field)
		}
		seen[field.Name] = true
	}
	var actual map[string]any
	if json.Unmarshal(delivery.Actual, &actual) != nil {
		t.Fatal("typed output missing")
	}
	want, _ := json.Marshal(expected)
	got, _ := json.Marshal(actual)
	if !bytes.Equal(got, want) || actual["zone"] != zone || len(actual["canonical"].(string)) != 1030 {
		t.Fatal("original thirteen-field contract or unchanged zone differs", actual)
	}
}

func assertTextBoundNative1905(t *testing.T, receipt packageExecutionReceipt, code, cases int) {
	t.Helper()
	if code != exitOK || receipt.Result == nil {
		t.Fatal("public package execution failed", code, receipt.Error)
	}
	r := receipt.Result.Runtime
	if !r.ProjectionReplayed || !r.RuntimeReplayed || r.FiniteTotal != cases || r.FinitePassed != cases ||
		r.ModelCalls != 0 || len(r.Runs) != 2 || len(r.Traces) != cases {
		t.Fatal("native typed comparison/replay did not complete", r)
	}
	if head := os.Getenv("GOOO_CI_SOURCE_HEAD"); head != "" && r.ProducerSourceSHA != head {
		t.Fatal("native producer is not the original CI source", r.ProducerSourceSHA, head)
	}
	for _, process := range r.Runs {
		if !process.Started || !process.Completed || process.ExitCode != 0 || process.TimedOut || process.Canceled {
			t.Fatal("native execution was incomplete", process)
		}
	}
}
