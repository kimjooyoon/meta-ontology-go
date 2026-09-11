package main

import (
	"strings"
	"testing"
)

func TestParseVerifierPackageSummariesPreservesBoundedRows(t *testing.T) {
	stdout := strings.Join([]string{
		"ok   github.com/kimjooyoon/meta-ontology-go/internal/meta 0.123s",
		"ok   github.com/kimjooyoon/meta-ontology-go/internal/cache (cached)",
		"?    github.com/kimjooyoon/meta-ontology-go/examples/demo [no test files]",
		"FAIL github.com/kimjooyoon/meta-ontology-go/internal/broken 2ms",
	}, "\n")

	parsed := parseVerifierPackageSummaries([]byte(stdout))
	if parsed.status() != "COMPLETE" || len(parsed.Rows) != 4 {
		t.Fatalf("parsed package summaries = %#v", parsed)
	}
	if parsed.Rows[0].ElapsedToken != "0.123s" || parsed.Rows[0].ElapsedStatus != "PARSED_EXACTLY" || parsed.Rows[0].ElapsedUnit != "NANOSECONDS" || parsed.Rows[0].ElapsedNanoseconds == nil || *parsed.Rows[0].ElapsedNanoseconds != 123000000 {
		t.Fatalf("exact elapsed row = %#v", parsed.Rows[0])
	}
	if parsed.Rows[1].OutputMarker != "OBSERVED_OUTPUT_MARKER" || parsed.Rows[1].ElapsedStatus != "UNKNOWN" {
		t.Fatalf("cached marker row = %#v", parsed.Rows[1])
	}
	if parsed.Rows[2].Status != "?" || parsed.Rows[2].OutputMarker != "OBSERVED_OUTPUT_MARKER" {
		t.Fatalf("no-test-files row = %#v", parsed.Rows[2])
	}
	if parsed.Rows[3].ElapsedToken != "2ms" || parsed.Rows[3].ElapsedNanoseconds == nil || *parsed.Rows[3].ElapsedNanoseconds != 2000000 {
		t.Fatalf("failure duration row = %#v", parsed.Rows[3])
	}
}

func TestParseVerifierPackageSummariesRejectsModulePrefixSpoofing(t *testing.T) {
	stdout := strings.Join([]string{
		"ok github.com/kimjooyoon/meta-ontology-go.evil/pkg 1s",
		"ok github.com/kimjooyoon/meta-ontology-go/../evil 1s",
		"ok github.com/kimjooyoon/meta-ontology-go/internal/valid 1s",
	}, "\n")

	parsed := parseVerifierPackageSummaries([]byte(stdout))
	if parsed.status() != "UNRECOGNIZED_INPUT" || parsed.UnrecognizedLineCount != 2 || len(parsed.Rows) != 1 {
		t.Fatalf("module boundary result = %#v", parsed)
	}
	if parsed.Rows[0].Package != "github.com/kimjooyoon/meta-ontology-go/internal/valid" {
		t.Fatalf("accepted package = %#v", parsed.Rows[0])
	}
}

func TestParseVerifierPackageSummariesMarksGlobalBounds(t *testing.T) {
	line := "ok github.com/kimjooyoon/meta-ontology-go/internal/valid 1s\n"
	parsed := parseVerifierPackageSummaries([]byte(strings.Repeat(line, maxVerifierPackageSummaryRowsPerInvocation+1)))
	if !parsed.Truncated || parsed.status() != "TRUNCATED" || len(parsed.Rows) != maxVerifierPackageSummaryRowsPerInvocation {
		t.Fatalf("row bound result = %#v", parsed)
	}

	longLine := "ok github.com/kimjooyoon/meta-ontology-go/internal/valid " + strings.Repeat("x", maxVerifierPackageSummaryLineBytes)
	parsed = parseVerifierPackageSummaries([]byte(longLine))
	if !parsed.Truncated || parsed.status() != "TRUNCATED" || len(parsed.Rows) != 0 {
		t.Fatalf("line bound result = %#v", parsed)
	}
}
