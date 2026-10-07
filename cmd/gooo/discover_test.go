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
		receipt.ProfileID != "gooo-jev-capability-discovery/v1" || receipt.FirstUnresolved == nil || receipt.FirstUnresolved.ID != "declaration_coverage" ||
		receipt.AggregateCompletenessScore != nil {
		t.Fatalf("discovery overstated its evidence: report=%s receipt=%#v", firstOut.String(), receipt)
	}
	if dimensionStatus(receipt, "declaration_coverage") != "UNKNOWN" || dimensionStatusDenominator(receipt, "declaration_coverage") != 0 ||
		dimensionStatus(receipt, "catalog_match") != "PASS" || dimensionStatus(receipt, "real_use_case_coverage") != "UNKNOWN" ||
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
	if receipt.Decision != "PROGRESS" || receipt.FirstUnresolved == nil || receipt.FirstUnresolved.ID != "declaration_coverage" ||
		dimensionStatus(receipt, "catalog_match") != "UNKNOWN" || dimensionStatus(receipt, "declaration_coverage") != "UNKNOWN" {
		t.Fatalf("unknown catalog query was reported as success: %#v", receipt)
	}
}

func TestDiscoverMeasuresDeclarationCoverageAgainstSeparateGoooContract(t *testing.T) {
	const source = `package billing
namespace billing
entity Bill id "billing://invoice"
entity Receipt id "billing://receipt"
activity Issue(Bill) -> Receipt computes "billing.issue"
`
	const contract = `package billing_contract
namespace billing
entity Invoice id "billing://invoice"
entity Receipt id "billing://receipt"
entity Reviewer id "billing://reviewer"
activity Issue(Invoice) -> Receipt
`
	reader := runSourceReaderWithFiles{
		"main.gooo":   []byte(source),
		"domain.gooo": []byte(contract),
	}
	var stdout, stderr bytes.Buffer
	code := runDiscover([]string{"--json", "--query", "How can I issue a billing receipt?", "--domain-contract", "domain.gooo", "main.gooo"},
		reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("discover code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report capabilityDiscoveryReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.DomainContractPath != "domain.gooo" || !strings.HasPrefix(report.DomainContractHash, "sha256:") || report.DomainContractSemantic == "" {
		t.Fatalf("domain contract identity is missing from report: %#v", report)
	}
	if err := completeness.Validate(report.Receipt); err != nil {
		t.Fatalf("validate scoped receipt: %v", err)
	}
	coverage := completenessDimension(report.Receipt, "declaration_coverage")
	if coverage.Status != "PROGRESS" || coverage.Numerator != 3 || coverage.Denominator != 4 ||
		!strings.Contains(strings.Join(coverage.Evidence, "\n"), "billing://reviewer") {
		t.Fatalf("declaration coverage did not use the independent contract denominator: %#v", coverage)
	}
	if !strings.Contains(strings.Join(report.Receipt.Dimensions[len(report.Receipt.Dimensions)-1].Evidence, "\n"), report.DomainContractHash) {
		t.Fatalf("provenance dimension did not bind the domain contract: %#v", report.Receipt.Dimensions[len(report.Receipt.Dimensions)-1])
	}

	reader["main.gooo"] = []byte(strings.Replace(source,
		"activity Issue(Bill) -> Receipt computes \"billing.issue\"",
		"activity Issue(Receipt) -> Receipt computes \"billing.issue\"", 1))
	stdout.Reset()
	stderr.Reset()
	code = runDiscover([]string{"--json", "--query", "How can I issue a billing receipt?", "--domain-contract", "domain.gooo", "main.gooo"},
		reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("typed-port mismatch discover code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	coverage = completenessDimension(report.Receipt, "declaration_coverage")
	if coverage.Status != "PROGRESS" || coverage.Numerator != 2 || coverage.Denominator != 4 {
		t.Fatalf("incompatible activity ports were counted as covered: %#v", coverage)
	}
}

func TestDiscoverPreservesOrderedTypedActivityInputsInDeclarationCoverage(t *testing.T) {
	const source = `package billing
namespace billing
entity Integer id "billing://integer"
entity Text id "billing://text"
activity Merge(Integer, Text) -> Text computes "return input1"
`
	const contract = `package billing_contract
namespace billing
entity Integer id "billing://integer"
entity Text id "billing://text"
activity Merge(Text, Integer) -> Text
`
	reader := runSourceReaderWithFiles{
		"main.gooo":   []byte(source),
		"domain.gooo": []byte(contract),
	}
	var stdout, stderr bytes.Buffer
	code := runDiscover([]string{"--json", "--query", "How do I merge a billing value?", "--domain-contract", "domain.gooo", "main.gooo"},
		reader, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("discover code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var report capabilityDiscoveryReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	coverage := completenessDimension(report.Receipt, "declaration_coverage")
	if coverage.Status != "PROGRESS" || coverage.Numerator != 2 || coverage.Denominator != 3 {
		t.Fatalf("reversed typed ports were counted as the same declaration: %#v", coverage)
	}
	evidence := strings.Join(coverage.Evidence, "\n")
	if !strings.Contains(evidence, `input_port_order_mismatches:[{"contract_input_order":["billing://text","billing://integer"]`) ||
		!strings.Contains(evidence, `"source_input_order":["billing://integer","billing://text"]`) {
		t.Fatalf("receipt did not explain the ordered-port mismatch: %q", evidence)
	}
}

func dimensionStatus(receipt *completeness.CompletenessReceipt, id string) string {
	return completenessDimension(receipt, id).Status
}

func dimensionStatusDenominator(receipt *completeness.CompletenessReceipt, id string) int {
	return completenessDimension(receipt, id).Denominator
}

func completenessDimension(receipt *completeness.CompletenessReceipt, id string) completeness.CompletenessDimension {
	for _, dimension := range receipt.Dimensions {
		if dimension.ID == id {
			return dimension
		}
	}
	return completeness.CompletenessDimension{ID: id, Status: "MISSING"}
}
