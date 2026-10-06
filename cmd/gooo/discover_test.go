package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

const discoverFixture = `package billing
namespace billing
entity Integer id "billing://integer"
activity Clamp(Integer) -> Integer computes "return input"
`

func TestDiscoverEmitsJEVTrailAndSharedCompletenessReceipt(t *testing.T) {
	var firstOut, firstErr bytes.Buffer
	code := runDiscover([]string{"--json", "--query", "How do I generate a canonical .gooo declaration?", "main.gooo"},
		runSourceReader{discoverFixture}, &firstOut, &firstErr)
	if code != exitOK || firstErr.Len() != 0 {
		t.Fatalf("discover code=%d stderr=%q stdout=%q", code, firstErr.String(), firstOut.String())
	}
	var first struct {
		Decision string          `json:"decision"`
		Trail    json.RawMessage `json:"capability_trail"`
		Source   string          `json:"source_digest"`
		Semantic string          `json:"semantic_digest"`
		Receipt  json.RawMessage `json:"completeness_receipt"`
	}
	if err := json.Unmarshal(firstOut.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	receipt, err := completeness.Decode(first.Receipt)
	if err != nil {
		t.Fatalf("decode shared completeness receipt: %v", err)
	}
	if err := completeness.Validate(receipt); err != nil {
		t.Fatalf("validate shared completeness receipt: %v", err)
	}
	if first.Decision != "PROGRESS" || len(first.Trail) == 0 || !strings.HasPrefix(first.Source, "sha256:") || first.Semantic == "" ||
		receipt.ProfileID != "gooo-jev-capability-discovery/v1" || receipt.FirstUnresolved == nil || receipt.FirstUnresolved.ID != "generation_coverage" ||
		receipt.AggregateCompletenessScore != nil {
		t.Fatalf("discovery overstated its evidence: report=%s receipt=%#v", firstOut.String(), receipt)
	}
	if dimensionStatus(receipt, "catalog_match") != "PASS" || dimensionStatus(receipt, "real_use_case_coverage") != "UNKNOWN" ||
		dimensionStatus(receipt, "reverse_observation_coverage") != "UNKNOWN" {
		t.Fatalf("catalog discovery was conflated with behavior evidence: %#v", receipt.Dimensions)
	}
	var secondOut, secondErr bytes.Buffer
	if code := runDiscover([]string{"--json", "--query", "How do I generate a canonical .gooo declaration?", "main.gooo"},
		runSourceReader{discoverFixture}, &secondOut, &secondErr); code != exitOK || secondErr.Len() != 0 || !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) {
		t.Fatalf("discovery was not deterministic: code=%d stderr=%q first=%s second=%s", code, secondErr.String(), firstOut.String(), secondOut.String())
	}
}

func TestDiscoverKeepsUnknownCatalogResultsOutOfSuccessMetrics(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runDiscover([]string{"--json", "--query", "Can Gooo explain quantum breakfast?", "main.gooo"},
		runSourceReader{discoverFixture}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("discover code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report struct {
		Receipt json.RawMessage `json:"completeness_receipt"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	receipt, err := completeness.Decode(report.Receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := completeness.Validate(receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PROGRESS" || receipt.FirstUnresolved == nil || receipt.FirstUnresolved.ID != "catalog_match" ||
		dimensionStatus(receipt, "catalog_match") != "UNKNOWN" {
		t.Fatalf("unknown catalog query was reported as success: %#v", receipt)
	}
}

func dimensionStatus(receipt *completeness.CompletenessReceipt, id string) string {
	for _, dimension := range receipt.Dimensions {
		if dimension.ID == id {
			return dimension.Status
		}
	}
	return "MISSING"
}
