package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceimpact"
)

const impactCLISource = `package sample
namespace sample
entity Integer id "sample://integer"
activity Reference(Integer) -> Integer id "sample://reference" computes "return input - 1"
activity Main(Integer) -> Integer id "sample://main" computes "return Reference(input)"
activity Independent(Integer) -> Integer id "sample://independent" computes "return input"
`

func impactCLIReader() mapSourceReader {
	return mapSourceReader{"before.gooo": []byte(impactCLISource), "after.gooo": []byte(strings.Replace(impactCLISource, "input - 1", "input - 2", 1))}
}

func TestSourceImpactCLITextAndJSON(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		var out, diagnostics bytes.Buffer
		if runSourceImpact([]string{"before.gooo", "after.gooo", "--format", format}, impactCLIReader(), &out, &diagnostics) != exitOK || diagnostics.Len() != 0 {
			t.Fatal(diagnostics.String())
		}
		if format == "json" {
			var report sourceimpact.Report
			if err := json.Unmarshal(out.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.Schema != sourceimpact.Schema || len(report.Affected) != 2 || report.ModelCalls != 0 || report.CandidateExecutions != 0 {
				t.Fatal(report)
			}
		} else if !strings.Contains(out.String(), "Recheck Main [sample://main]") || strings.Contains(out.String(), "Recheck Independent") || !strings.Contains(out.String(), "Model calls: 0; candidate executions: 0") {
			t.Fatal(out.String())
		}
	}
}

func TestSourceImpactCLIRejectsBeforeProducingReport(t *testing.T) {
	for _, args := range [][]string{nil, {"one"}, {"one", "two", "three"}, {"one", "two", "--format", "csv"}, {"--bad", "two"}} {
		var out, diagnostics bytes.Buffer
		if runSourceImpact(args, impactCLIReader(), &out, &diagnostics) != exitUsage || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal(args, &out, &diagnostics)
		}
	}
	for _, broken := range []mapSourceReader{{"before.gooo": []byte(impactCLISource)}, {"before.gooo": []byte(impactCLISource), "after.gooo": []byte("invalid")}} {
		var out, diagnostics bytes.Buffer
		if runSourceImpact([]string{"before.gooo", "after.gooo"}, broken, &out, &diagnostics) != exitFailure || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal(&out, &diagnostics)
		}
	}
}

func TestSourceImpactCLIReportsPartialWithoutClaimingFailure(t *testing.T) {
	reader := impactCLIReader()
	reader["after.gooo"] = []byte(strings.Replace(impactCLISource, "Reference(input)", "external.Reference(input)", 1))
	var out, diagnostics bytes.Buffer
	if runSourceImpact([]string{"before.gooo", "after.gooo"}, reader, &out, &diagnostics) != exitOK || !strings.Contains(out.String(), "PARTIAL_STATIC") || !strings.Contains(out.String(), "Unresolved after") || !strings.Contains(out.String(), "external.Reference") {
		t.Fatal(&out, &diagnostics)
	}
}

type impactFailedWriter struct{}

func (impactFailedWriter) Write([]byte) (int, error) { return 0, errors.New("destination closed") }

func TestSourceImpactCLIWriterErrorAndPublicDispatch(t *testing.T) {
	var diagnostics bytes.Buffer
	if runSourceImpact([]string{"before.gooo", "after.gooo"}, impactCLIReader(), impactFailedWriter{}, &diagnostics) != exitFailure || !strings.Contains(diagnostics.String(), "destination closed") {
		t.Fatal(&diagnostics)
	}
	dir := t.TempDir()
	for name, contents := range impactCLIReader() {
		if err := os.WriteFile(filepath.Join(dir, name), contents, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	diagnostics.Reset()
	if run([]string{"impact", filepath.Join(dir, "before.gooo"), filepath.Join(dir, "after.gooo")}, &out, &diagnostics) != exitOK || !strings.Contains(out.String(), "COMPLETE_STATIC") {
		t.Fatal(&out, &diagnostics)
	}
	for _, args := range [][]string{{"help", "impact"}, {"impact", "--help"}, {"help", "commands"}, {"--help"}} {
		out.Reset()
		diagnostics.Reset()
		if run(args, &out, &diagnostics) != exitOK || !strings.Contains(out.String(), "impact") {
			t.Fatal(args, &out, &diagnostics)
		}
	}
}
