package metricprogram_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

func TestCompileBindsEveryIndicatorToVerifiedGoooMetaCode(t *testing.T) {
	strategy, strategyVerification := fixturePayloads(t)
	program, source, err := metricprogram.Compile(strategy, strategyVerification)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../../examples/metric-meta-program/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, fixture) {
		t.Fatal("compiled meta source differs from the checked-in Gooo fixed point")
	}
	if program.Coverage.Status != "COMPLETE" || program.Coverage.BindingCount != 15 || program.Coverage.ResolvedBindingCount != 15 || program.Coverage.RegistryOperationCount != 8 || program.Coverage.ReferencedOperationCount != 8 {
		t.Fatalf("coverage = %#v", program.Coverage)
	}
	if len(program.Steps) != 4 || program.Steps[0].OperationID != "observe-counterfactual-boundary" || program.Steps[3].OperationID != "terminate-at-fixed-point" {
		t.Fatalf("steps = %#v", program.Steps)
	}
	programPayload := fixtureJSON(t, program)
	report, err := programverify.Verify(strategy, strategyVerification, programPayload, source)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "VERIFIED" || report.BindingCount != 15 || report.OperationCount != 8 || report.StepCount != 4 || report.RepositoryWorkspaceWrites || report.PromotionAuthorized {
		t.Fatalf("verification = %#v", report)
	}
}

func TestCompileIsDeterministic(t *testing.T) {
	strategy, verification := fixturePayloads(t)
	first, firstSource, err := metricprogram.Compile(strategy, verification)
	if err != nil {
		t.Fatal(err)
	}
	second, secondSource, err := metricprogram.Compile(strategy, verification)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest != second.Digest || !bytes.Equal(firstSource, secondSource) {
		t.Fatal("meta program compilation is not deterministic")
	}
}

func TestCompileRejectsUnboundOperationAndRootReadmePolicy(t *testing.T) {
	strategyPayload, verification := fixturePayloads(t)
	var strategy metricprogram.StrategyPlan
	if err := json.Unmarshal(strategyPayload, &strategy); err != nil {
		t.Fatal(err)
	}
	strategy.Bindings[0].MetaOperation = "rewrite-repository"
	if _, _, err := metricprogram.Compile(fixtureJSON(t, strategy), verification); err == nil || !strings.Contains(err.Error(), "not resolved") {
		t.Fatalf("unknown operation error = %v", err)
	}
	strategyPayload, verification = fixturePayloads(t)
	if err := json.Unmarshal(strategyPayload, &strategy); err != nil {
		t.Fatal(err)
	}
	strategy.RootPolicy.ReadmeRequirement = "REQUIRED"
	if _, _, err := metricprogram.Compile(fixtureJSON(t, strategy), verification); err == nil || !strings.Contains(err.Error(), "root exception") {
		t.Fatalf("root policy error = %v", err)
	}
}

func TestIndependentVerifierRejectsTamperedProgramAndSource(t *testing.T) {
	strategy, verification := fixturePayloads(t)
	program, source, err := metricprogram.Compile(strategy, verification)
	if err != nil {
		t.Fatal(err)
	}
	program.Steps[0].Mode = "WRITE"
	if _, err := programverify.Verify(strategy, verification, fixtureJSON(t, program), source); err == nil || !strings.Contains(err.Error(), "independent reconstruction") {
		t.Fatalf("tampered program error = %v", err)
	}
	program, source, err = metricprogram.Compile(strategy, verification)
	if err != nil {
		t.Fatal(err)
	}
	source = append(append([]byte(nil), source...), '\n')
	if _, err := programverify.Verify(strategy, verification, fixtureJSON(t, program), source); err == nil || !strings.Contains(err.Error(), "not canonical") {
		t.Fatalf("tampered source error = %v", err)
	}
}
