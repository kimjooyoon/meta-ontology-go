package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func TestDiscoverExecutesGenerationAndCountsUniqueIndependentInputs(t *testing.T) {
	for _, fixture := range []struct{ source, activity string }{
		{"ir-search-source", "ClampNegativeToZero"}, {"source-assembly", "Clamp"},
	} {
		t.Run(fixture.source, func(t *testing.T) {
			reader := discoverGenerationFixture(t, fixture.source, fixture.activity)
			original := bytes.Clone(reader["generation.json"])
			reader["cases.json"] = []byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[
{"input":-5,"expected":0},{"input":5,"expected":5},{"input":-5,"expected":0},{"input":0,"expected":0}]}`)
			result := executeDiscoveryFixture(t, reader, discoveryTestGoBinary(), exitOK)
			d := completenessDimension(result.Receipt, "real_use_case_coverage")
			if d.Status != "PASS" || d.Numerator != 2 || d.Denominator != 2 {
				t.Fatal("repeated or selection inputs increased independent coverage", d)
			}
			if result.Runtime == nil || result.Runtime.Observation.Stage != "COMPLETE" || !result.Runtime.Observation.RuntimeReplayed || len(result.Runtime.Observation.Runs) != 2 {
				t.Fatal("missing fresh native replay", result.Runtime)
			}
			for _, id := range []string{"reverse_observation_coverage", "execution_boundary_coverage", "runtime_finite_accuracy"} {
				if dimensionStatus(result.Receipt, id) != "PASS" {
					t.Fatal("native observation did not reach the discovery receipt", id)
				}
			}
			if dimensionStatus(result.Receipt, "permission_boundary_coverage") != "UNKNOWN" ||
				dimensionStatus(result.Receipt, "generation_coverage") != "UNKNOWN" || result.Decision != "PROGRESS" {
				t.Fatal("native cases silently closed a different obligation", result.Receipt)
			}
			if !bytes.Equal(original, reader["generation.json"]) {
				t.Fatal("runtime observation changed its parent generation")
			}
			wire, err := json.Marshal(result.Runtime)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := bodyexecution.DecodeRuntimeReceipt(wire); err != nil {
				t.Fatal("nested runtime lost its parent binding", err)
			}
		})
	}
}

func TestDiscoverRuntimePreservesZeroMatchesAndSelectionOnlyCases(t *testing.T) {
	for _, fixture := range []struct {
		name, cases, status string
		denominator         int
	}{
		{"conflicting-duplicate", `[{"input":-5,"expected":0},{"input":-5,"expected":99},{"input":0,"expected":0}]`, "PROGRESS", 1},
		{"selection-only", `[{"input":-2,"expected":0},{"input":0,"expected":0}]`, "UNKNOWN", 0},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			reader := discoverGenerationFixture(t, "ir-search-source", "ClampNegativeToZero")
			reader["cases.json"] = []byte(`{"schema":"gooo/body-runtime-cases/v1","cases":` + fixture.cases + `}`)
			result := executeDiscoveryFixture(t, reader, discoveryTestGoBinary(), exitOK)
			d := completenessDimension(result.Receipt, "real_use_case_coverage")
			if d.Status != fixture.status || d.Numerator != 0 || d.Denominator != fixture.denominator {
				t.Fatal("zero matching or selection-only inputs lost their measurement state", d)
			}
		})
	}
}

func TestDiscoverRuntimeRecordsToolFailureAndRequiresExplicitInputs(t *testing.T) {
	reader := discoverGenerationFixture(t, "ir-search-source", "ClampNegativeToZero")
	reader["cases.json"] = []byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":5,"expected":5}]}`)
	result := executeDiscoveryFixture(t, reader, t.TempDir(), exitFailure)
	if result.Decision != "FAIL_CLOSED" || result.Receipt.FailClosedReason == nil || result.Runtime == nil ||
		result.Runtime.Observation.Failure == "" || len(result.Runtime.Observation.Runs) != 0 ||
		dimensionStatus(result.Receipt, "real_use_case_coverage") != "UNKNOWN" {
		t.Fatal("failed execution became a runtime success", result)
	}
	for _, flags := range [][]string{
		{"--execute-cases", "cases.json"}, {"--go-bin", discoveryTestGoBinary()},
		{"--generation", "generation.json", "--execute-cases", "cases.json", "--execute-cases", "cases.json"},
	} {
		var stdout, stderr bytes.Buffer
		args := append([]string{"--json", "--query", "Generate Gooo code", "main.gooo"}, flags...)
		if code := runDiscover(args, reader, &stdout, &stderr); code != exitUsage {
			t.Fatal("invalid execution flags were accepted", flags, code, stdout.String(), stderr.String())
		}
	}
}

func executeDiscoveryFixture(t *testing.T, reader SourceReader, goBinary string, wantCode int) capabilityDiscoveryReport {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runDiscover([]string{"--json", "--query", "Generate Gooo code", "main.gooo", "--generation", "generation.json",
		"--execute-cases", "cases.json", "--go-bin", goBinary}, reader, &stdout, &stderr)
	if code != wantCode {
		t.Fatalf("discovery returned %d, want %d: %s %s", code, wantCode, stderr.String(), stdout.String())
	}
	var report capabilityDiscoveryReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if err := completeness.Validate(report.Receipt); err != nil {
		t.Fatal("invalid discovery runtime receipt", err, stdout.String())
	}
	return report
}

func discoveryTestGoBinary() string {
	path := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	return path
}
