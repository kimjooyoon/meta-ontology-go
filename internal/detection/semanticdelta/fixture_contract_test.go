package semanticdelta

import (
	"bytes"
	"embed"
	"reflect"
	"testing"
)

// The experiment is intentionally falsifiable: a single out-of-scope endpoint
// must change the report from pass to fail without changing the input contract.
const (
	hypothesis   = "H1: canonical scope detection is permutation-invariant and rejects every out-of-scope delta endpoint"
	passRule     = "pass when canonical Report is allowed with zero violations"
	failRule     = "fail when the exact out-of-scope endpoint violation is present"
	deferredRule = "defer host promotion; this fixture does not execute a gooo-hosted compiler"
)

//go:embed testdata/*.json testdata/*.txt
var fixtureFiles embed.FS

func TestScopeHypothesisFixturePassesAcrossEncodings(t *testing.T) {
	t.Logf("hypothesis=%s pass=%s fail=%s deferred=%s", hypothesis, passRule, failRule, deferredRule)
	jsonRequest := loadFixtureRequest(t, "hypothesis-in-scope.json")
	textRequest := loadFixtureRequest(t, "hypothesis-in-scope.txt")
	jsonReport := detectFixture(t, jsonRequest)
	textReport := detectFixture(t, textRequest)
	if !jsonReport.Passes() || !textReport.Passes() {
		t.Fatalf("in-scope fixture failed: JSON=%#v text=%#v", jsonReport, textReport)
	}
	if !reflect.DeepEqual(jsonReport, textReport) {
		t.Fatalf("JSON/text reports diverged: %#v != %#v", jsonReport, textReport)
	}
	permuted := permuteRequest(jsonRequest)
	canonical, err := EncodeJSON(jsonRequest)
	if err != nil {
		t.Fatal(err)
	}
	permutedCanonical, err := EncodeJSON(permuted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(canonical, permutedCanonical) {
		t.Fatalf("presentation/order changed canonical request:\n%s\n%s", canonical, permutedCanonical)
	}
}

func TestScopeHypothesisCounterexampleFailsWithOneEndpointViolation(t *testing.T) {
	request := loadFixtureRequest(t, "counterexample-out-of-scope.json")
	report := detectFixture(t, request)
	want := Report{Allowed: false, Violations: []Violation{{
		Operation: OperationAdd, Change: ChangeFact,
		Subject: "billing://activity/pay-order", Predicate: "prov:used",
		Object: "fraud://entity/charge", Endpoint: "object",
		Reason: "fact object is outside allowed scope",
	}}}
	if !reflect.DeepEqual(report, want) {
		t.Fatalf("counterexample report = %#v, want %#v", report, want)
	}
}

func TestScopeReportOutputMatchesCanonicalFixtures(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{input: "hypothesis-in-scope.json", output: "expected-in-scope-report.json"},
		{input: "counterexample-out-of-scope.json", output: "expected-counterexample-report.json"},
	}
	for _, testCase := range cases {
		t.Run(testCase.input, func(t *testing.T) {
			request := loadFixtureRequest(t, testCase.input)
			report := detectFixture(t, request)
			actual, err := EncodeReportJSON(report)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := fixtureFiles.ReadFile("testdata/" + testCase.output)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expected) {
				t.Fatalf("canonical report mismatch:\nactual=%s\nexpected=%s", actual, expected)
			}
		})
	}
}

func loadFixtureRequest(t *testing.T, name string) Request {
	t.Helper()
	data, err := fixtureFiles.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	request, err := Decode(data)
	if err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return request
}

func detectFixture(t *testing.T, request Request) Report {
	t.Helper()
	report, err := Detect(request.Delta, request.Allowed)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func permuteRequest(request Request) Request {
	request.Allowed.IDs = reverseStrings(request.Allowed.IDs)
	request.Allowed.Predicates = reverseStrings(request.Allowed.Predicates)
	request.Delta.AddedNodes = reverseNodes(request.Delta.AddedNodes)
	request.Delta.AddedFacts = reverseFacts(request.Delta.AddedFacts)
	return request
}

func reverseStrings(values []string) []string {
	result := append([]string(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseNodes(values []Node) []Node {
	result := append([]Node(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseFacts(values []Fact) []Fact {
	result := append([]Fact(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
