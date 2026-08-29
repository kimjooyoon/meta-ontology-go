package metricprogram_test

import (
	"bytes"
	"os"
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
	if program.Coverage.Status != "COMPLETE" || program.Coverage.BindingCount != 16 || program.Coverage.ResolvedBindingCount != 16 || program.Coverage.RegistryOperationCount != 9 || program.Coverage.ReferencedOperationCount != 9 {
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
	if report.Status != "VERIFIED" || report.BindingCount != 16 || report.OperationCount != 9 || report.StepCount != 4 || report.RepositoryWorkspaceWrites || report.PromotionAuthorized {
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

func TestCompileSeparatesNonPromotingTerminalPath(t *testing.T) {
	strategy, verification := fixturePayloads(t)
	var plan metricprogram.StrategyPlan
	if err := json.Unmarshal(strategy, &plan); err != nil {
		t.Fatal(err)
	}
	plan.Selection.Decision = "PRESERVE_NON_PROMOTING_TERMINAL"
	plan.Selection.MetaOperation = "preserve-non-promoting-terminal"
	plan.Selection.Reason = "NON_PROMOTING_TERMINAL_PRESERVED"
	strategy = fixtureJSON(t, plan)
	program, source, err := metricprogram.Compile(strategy, verification)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Steps) != 4 || program.Steps[2].OperationID != "replay-counterfactual" ||
		program.Steps[3].OperationID != "preserve-non-promoting-terminal" ||
		program.Steps[3].OutputEntity != "NonPromotingTerminalReceipt" {
		t.Fatalf("steps = %#v", program.Steps)
	}
	for _, step := range program.Steps {
		if step.OperationID == "terminate-at-fixed-point" {
			t.Fatalf("mixed path contains fixed-point terminator: %#v", program.Steps)
		}
	}
	if _, err := programverify.Verify(strategy, verification, fixtureJSON(t, program), source); err != nil {
		t.Fatal(err)
	}
}
