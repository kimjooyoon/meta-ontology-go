package linecapscontract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultBoundaryIsMeasuredAndPasses(t *testing.T) {
	source := exactDefaultSource()
	boundary := Case{
		ID: "default-boundary", Hypothesis: "exact caps pass", Fixture: "generated/exact.go",
		Limits: DefaultLimits(), Expected: Pass,
	}
	evidence, err := EvaluateCase(boundary, "generated/exact.go", source)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Decision != Pass || !evidence.OutcomeMatches || evidence.Measurement.FileLines != 300 {
		t.Fatalf("unexpected boundary evidence: %#v", evidence)
	}
	if len(evidence.Measurement.Functions) != 1 || evidence.Measurement.Functions[0].Lines != 75 {
		t.Fatalf("function boundary was not measured: %#v", evidence.Measurement.Functions)
	}
	overFile := append(append([]byte(nil), source...), '\n')
	boundary.Expected = Fail
	overEvidence, err := EvaluateCase(boundary, "generated/exact.go", overFile)
	if err != nil {
		t.Fatal(err)
	}
	if overEvidence.Decision != Fail || !overEvidence.OutcomeMatches || overEvidence.Findings[0].Rule != RuleFileLines || overEvidence.Findings[0].Actual != 301 {
		t.Fatalf("one line over file boundary was not rejected: %#v", overEvidence)
	}
}

func TestFixtureCounterexamplesAndDeferredCase(t *testing.T) {
	sources := loadFixtures(t, "safe.go.txt", "over.go.txt", "invalid.go.txt", "deferred.go.txt")
	cases := []Case{
		{ID: "deferred-host", Hypothesis: "future host parity is not inferred from Go AST", Fixture: "deferred.go.txt", Limits: Limits{MaxFileLines: 3, MaxFunctionLines: 3}, Expected: Deferred, DeferredReason: "gooo-hosted parity is not implemented"},
		{ID: "invalid-syntax", Hypothesis: "unparseable input cannot pass verification", Fixture: "invalid.go.txt", Limits: Limits{MaxFileLines: 10, MaxFunctionLines: 10}, Expected: Fail},
		{ID: "over-function", Hypothesis: "one line over the function cap fails", Fixture: "over.go.txt", Limits: Limits{MaxFileLines: 10, MaxFunctionLines: 3}, Expected: Fail},
		{ID: "safe-source", Hypothesis: "a source at its local caps passes", Fixture: "safe.go.txt", Limits: Limits{MaxFileLines: 6, MaxFunctionLines: 4}, Expected: Pass},
	}
	evidence, err := EvaluateCases(cases, sources)
	if err != nil {
		t.Fatal(err)
	}
	if len(evidence) != len(cases) {
		t.Fatalf("unexpected evidence count: %d", len(evidence))
	}
	for _, result := range evidence {
		if !result.OutcomeMatches {
			t.Errorf("case %s did not meet its expected decision: %#v", result.CaseID, result)
		}
	}
	if evidence[0].CaseID != "deferred-host" || evidence[0].Decision != Deferred || evidence[0].Measurement.Parseable == false {
		t.Fatalf("deferred case was reported as success or parse failure: %#v", evidence[0])
	}
	if evidence[1].Decision != Fail || evidence[1].Findings[0].Rule != RuleParseFile {
		t.Fatalf("invalid case was not a parse failure: %#v", evidence[1])
	}
}

func TestLineEndingAndInputOrderHypothesis(t *testing.T) {
	source := []byte("package fixture\n\nfunc F() {\n\t_ = 1\n}\n")
	crlf := []byte(strings.ReplaceAll(string(source), "\n", "\r\n"))
	limits := Limits{MaxFileLines: 5, MaxFunctionLines: 3}
	left, err := Measure("left.go", source)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Measure("right.go", crlf)
	if err != nil {
		t.Fatal(err)
	}
	if left.FileLines != right.FileLines || left.Parseable != right.Parseable || left.Functions[0] != right.Functions[0] {
		t.Fatalf("line-ending normalization changed measurements: left=%#v right=%#v", left, right)
	}
	cases := []Case{
		{ID: "z-case", Hypothesis: "order does not affect evidence order", Fixture: "safe.go.txt", Limits: limits, Expected: Pass},
		{ID: "a-case", Hypothesis: "order does not affect evidence order", Fixture: "safe.go.txt", Limits: limits, Expected: Pass},
	}
	sources := map[string][]byte{"safe.go.txt": source}
	first, err := EvaluateCases(cases, sources)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EvaluateCases([]Case{cases[1], cases[0]}, sources)
	if err != nil {
		t.Fatal(err)
	}
	if first[0].CaseID != "a-case" || second[0].CaseID != first[0].CaseID || first[0].Measurement.SourceDigest != second[0].Measurement.SourceDigest {
		t.Fatalf("case order was not canonicalized: first=%#v second=%#v", first, second)
	}
	firstJSON, err := first[0].JSON()
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := second[0].JSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("case order changed evidence serialization: first=%s second=%s", firstJSON, secondJSON)
	}
}

func TestJSONIncludesContractAndDeferredState(t *testing.T) {
	evidence, err := EvaluateCase(Case{
		ID: "json", Hypothesis: "deferred is explicit", Fixture: "deferred.go.txt",
		Limits: Limits{MaxFileLines: 3, MaxFunctionLines: 3}, Expected: Deferred,
		DeferredReason: "future stage not implemented",
	}, "deferred.go.txt", loadFixture(t, "deferred.go.txt"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := evidence.JSON()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, fragment := range []string{`"schema": "gooo/linecaps-evidence/v1"`, `"decision": "deferred"`, `"outcome_matches": true`} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("JSON omitted %q: %s", fragment, text)
		}
	}
}

func exactDefaultSource() []byte {
	var source strings.Builder
	source.WriteString("package fixture\n\nfunc Exact() {\n")
	for i := 0; i < 73; i++ {
		source.WriteString("\t_ = 1\n")
	}
	source.WriteString("}\n")
	for i := 0; i < 223; i++ {
		source.WriteByte('\n')
	}
	return []byte(source.String())
}

func loadFixtures(t *testing.T, names ...string) map[string][]byte {
	t.Helper()
	fixtures := make(map[string][]byte, len(names))
	for _, name := range names {
		fixtures[name] = loadFixture(t, name)
	}
	return fixtures
}

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
