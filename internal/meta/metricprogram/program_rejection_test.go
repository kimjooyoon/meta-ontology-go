package metricprogram_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
)

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
