package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func TestDiscoverReplaysSavedGenerationAgainstDomainActivities(t *testing.T) {
	for _, fixture := range []struct{ source, activity, contract string }{
		{"ir-search-source", "ClampNegativeToZero", `package expected
namespace bodycodegen
entity Integer id "bodycodegen://entity/integer"
activity ClampNegativeToZero(Integer) -> Integer
activity Pending(Integer) -> Integer
`},
		{"source-assembly", "Clamp", `package expected
namespace source_assembly
entity Integer id "source-assembly://integer"
activity Clamp(Integer) -> Integer
activity Qualified(Integer) -> Integer
`},
	} {
		t.Run(fixture.source, func(t *testing.T) {
			checkDiscoveryGenerationDomain(t, fixture.source, fixture.activity, fixture.contract)
		})
	}
}

func TestDiscoverRejectsGenerationFromDifferentSourceOrChangedProjection(t *testing.T) {
	for _, change := range []string{"source", "projection", "finite-observation", "missing", "duplicate-flag"} {
		t.Run(change, func(t *testing.T) {
			checkDiscoveryChangedGeneration(t, change)
		})
	}
}

func discoverGenerationFixture(t *testing.T, name, activity string) runSourceReaderWithFiles {
	t.Helper()
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	filename := "../../examples/body-codegen/" + name + ".gooo.fixture"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", activity, filename}, &stdout, &stderr); code != exitOK {
		t.Fatalf("fixture generation failed: %d %s", code, stderr.String())
	}
	return runSourceReaderWithFiles{"main.gooo": source, "generation.json": stdout.Bytes()}
}

func runDiscoveryGenerationFixture(t *testing.T, reader SourceReader, contract bool) ([]byte, capabilityDiscoveryReport) {
	t.Helper()
	args := []string{"--json", "--query", "Generate Gooo code", "--generation", "generation.json", "main.gooo"}
	if contract {
		args = append(args, "--domain-contract", "domain.gooo")
	}
	var stdout, stderr bytes.Buffer
	if code := runDiscover(args, reader, &stdout, &stderr); code != exitOK {
		t.Fatalf("discover failed: %d %s %s", code, stderr.String(), stdout.String())
	}
	var result capabilityDiscoveryReport
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if err := completeness.Validate(result.Receipt); err != nil {
		t.Fatal(err)
	}
	return stdout.Bytes(), result
}

func checkDiscoveryGenerationDomain(t *testing.T, source, activity, contract string) {
	t.Helper()
	reader := discoverGenerationFixture(t, source, activity)
	reader["domain.gooo"] = []byte(contract)

	first, result := runDiscoveryGenerationFixture(t, reader, true)
	coverage := completenessDimension(result.Receipt, "generation_coverage")
	if coverage.Status != "PROGRESS" || coverage.Numerator != 1 || coverage.Denominator != 2 ||
		result.Generation == nil || !result.Generation.ProjectionReplay || result.Generation.Digest != "sha256:"+sha256Hex(reader["generation.json"]) {
		t.Fatalf("source generation evidence is missing or overstated: %#v %#v", coverage, result.Generation)
	}
	for _, id := range []string{"reverse_observation_coverage", "real_use_case_coverage"} {
		if dimensionStatus(result.Receipt, id) != "UNKNOWN" {
			t.Fatal("projection replay was counted as independent runtime evidence", id)
		}
	}
	second, _ := runDiscoveryGenerationFixture(t, reader, true)
	if !bytes.Equal(first, second) {
		t.Fatal("saved-generation discovery did not replay byte-for-byte")
	}
	_, unscoped := runDiscoveryGenerationFixture(t, reader, false)
	coverage = completenessDimension(unscoped.Receipt, "generation_coverage")
	if coverage.Status != "UNKNOWN" || coverage.Denominator != 0 || !unscoped.Generation.ProjectionReplay {
		t.Fatal("the artifact defined its own domain denominator", coverage)
	}
	reader["domain.gooo"] = []byte(contract[:strings.LastIndex(contract, "activity ")])
	_, single := runDiscoveryGenerationFixture(t, reader, true)
	if coverage := completenessDimension(single.Receipt, "generation_coverage"); coverage.Status != "PASS" || coverage.Numerator != 1 || coverage.Denominator != 1 {
		t.Fatal("replayed single-activity contract did not close generation coverage", coverage)
	}
	reader["domain.gooo"] = bytes.Replace(reader["domain.gooo"], []byte("(Integer)"), []byte("(Integer, Integer)"), 1)
	_, differentSignature := runDiscoveryGenerationFixture(t, reader, true)
	if coverage := completenessDimension(differentSignature.Receipt, "generation_coverage"); coverage.Status != "PROGRESS" || coverage.Numerator != 0 || coverage.Denominator != 1 {
		t.Fatal("projection for a different typed signature was counted", coverage)
	}
}

func checkDiscoveryChangedGeneration(t *testing.T, change string) {
	t.Helper()
	reader := discoverGenerationFixture(t, "ir-search-source", "ClampNegativeToZero")
	args := []string{"--json", "--query", "Generate Gooo code", "--generation", "generation.json", "main.gooo"}
	switch change {
	case "source":
		reader["main.gooo"] = append(reader["main.gooo"], '\n')
	case "missing":
		delete(reader, "generation.json")
	case "duplicate-flag":
		args = append(args, "--generation", "generation.json")
	default:
		var generated bodycodegen.Result
		if err := json.Unmarshal(reader["generation.json"], &generated); err != nil {
			t.Fatal(err)
		}
		if change == "projection" {
			generated.Source += "\n// changed\n"
		} else {
			generated.Report.BodySearch.TrainingCaseResults[0].Actual++
		}
		raw, err := json.Marshal(generated)
		if err != nil {
			t.Fatal(err)
		}
		reader["generation.json"] = raw
	}
	var stdout, stderr bytes.Buffer
	want := exitFailure
	if change == "duplicate-flag" {
		want = exitUsage
	}
	if code := runDiscover(args, reader, &stdout, &stderr); code != want {
		t.Fatalf("discovery returned %d for %s evidence: %s", code, change, stdout.String())
	}
	if change != "duplicate-flag" && !strings.Contains(stdout.String(), "GENERATION_REPLAY_FAILED") {
		t.Fatal("failure did not identify generation replay", stdout.String())
	}
}
