package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

type discoveryExecution struct {
	Result      bodyexecution.Result
	Independent completeness.CompletenessDimension
	Digest      string
}

func executeDiscoveryGeneration(reader SourceReader, filename string, source []byte,
	generation *discoveryGeneration, casesPath, goBinary string) (*discoveryExecution, error) {
	if casesPath == "" {
		return nil, nil
	}
	raw, err := reader.ReadFile(casesPath)
	if err != nil {
		return nil, err
	}
	cases, err := bodyexecution.DecodeCases(raw)
	if err != nil {
		return nil, err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	document, err := bodyexecution.DecodeSourcePlan(ctx, filename, source, generation.prior.Report.Activity, nil)
	if err != nil {
		return nil, err
	}
	result, runErr := bodyexecution.Execute(ctx, filename, source, document, generation.prior, generation.parent, cases, goBinary)
	wire, _ := json.Marshal(result)
	observation := &discoveryExecution{Result: result, Digest: "sha256:" + sha256Hex(wire)}
	observation.Independent = discoveryIndependentCases(document, generation, cases, result)
	return observation, runErr
}

func discoveryIndependentCases(document pathplan.Document, generation *discoveryGeneration,
	cases []pathplan.TestCase, result bodyexecution.Result) completeness.CompletenessDimension {
	unique := map[int64]bool{}
	for _, c := range cases {
		if !bodyexecution.SelectionObservedInput(document, generation.prior, c.Input) {
			unique[c.Input] = true
		}
	}
	d := completeness.CompletenessDimension{ID: "real_use_case_coverage", Status: "UNKNOWN", Denominator: len(unique),
		Unit:   "unique inputs for the selected activity, absent from its selection suite, with all supplied expectations matched in two native runs",
		Reason: "Fresh execution and at least one input absent from the effective selection suite are required.",
		Evidence: []string{"activity_id:" + generation.ActivityID, "runtime_suite_digest:" + result.Observation.RuntimeSuiteSHA256,
			"duplicate_inputs_count_once", "model_training_independence:unobserved"}}
	if result.Observation.Stage != "COMPLETE" || !result.Observation.RuntimeReplayed || len(unique) == 0 {
		return d
	}
	matched := map[int64]bool{}
	for _, c := range result.Observation.Cases {
		if unique[c.Input] {
			previous, observed := matched[c.Input]
			matched[c.Input] = c.Passed && (!observed || previous)
		}
	}
	for input := range unique {
		if matched[input] {
			d.Numerator++
		}
	}
	d.Status, d.Reason = "PROGRESS", "Some supplied expectations for independent inputs differ from the observed native outputs."
	if d.Numerator == d.Denominator {
		d.Status, d.Reason = "PASS", "All supplied expectations for these unique independent inputs matched in both native executions."
	}
	return d
}

func attachDiscoveryRuntime(receipt *completeness.CompletenessReceipt, execution *discoveryExecution) {
	runtime := execution.Result
	previousNext := map[string]string{}
	for _, claim := range receipt.UnresolvedClaims {
		previousNext[claim.ID] = claim.NextOperation
	}
	for _, dimension := range runtime.CompletenessReceipt.Dimensions {
		switch dimension.ID {
		case "reverse_observation_coverage", "runtime_completion", "runtime_finite_accuracy", "runtime_deterministic_replay", "runtime_child_resources":
			replaceDiscoveryDimension(receipt, dimension)
		case "execution_boundary":
			dimension.ID = "execution_boundary_coverage"
			replaceDiscoveryDimension(receipt, dimension)
		}
	}
	replaceDiscoveryDimension(receipt, execution.Independent)
	for i := range receipt.Dimensions {
		if receipt.Dimensions[i].ID == "provenance_integrity" {
			receipt.Dimensions[i].Evidence = append(receipt.Dimensions[i].Evidence, "fresh_runtime_digest:"+execution.Digest,
				"parent_receipt_digest:"+runtime.Observation.ParentReceiptSHA256)
		}
	}
	receipt.Scope["runtime_observation"] = map[string]any{"artifact_digest": execution.Digest,
		"runtime_scope": runtime.CompletenessReceipt.Scope["runtime_scope"], "boundary": runtime.CompletenessReceipt.Scope["boundary"]}
	receipt.Scope["investment_limit"] = map[string]any{"queries": 1, "model_calls": 0, "generation_attempts": 0,
		"native_builds": 1, "native_runs": 2}
	receipt.Scope["excluded_scope"] = []string{"behavior beyond the supplied cases", "model-training independence", "permission authorization", "external network behavior"}
	receipt.DecisionBasis = "The discovery trail, source-owned generation, and fresh native execution are linked. Each axis reports its declared denominator; independent input coverage applies to the selected activity."
	receipt.NotClaimed[2] = "Native observations cover the recorded finite inputs; model-training independence and whole-domain behavior remain unobserved."
	receipt.StatusCounts = map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	receipt.UnresolvedClaims = nil
	for _, d := range receipt.Dimensions {
		receipt.StatusCounts[d.Status]++
		if d.Status != "PASS" {
			next := previousNext[d.ID]
			if next == "" {
				next = nextCapabilityReceiptOperation(d.ID)
			}
			if d.ID == "real_use_case_coverage" && d.Denominator == 0 {
				next = "provide_unique_inputs_absent_from_the_selection_suite_with_expected_outputs"
			}
			receipt.UnresolvedClaims = append(receipt.UnresolvedClaims, completeness.UnresolvedCompletenessClaim{
				ID: d.ID, Status: d.Status, Reason: d.Reason, NextOperation: next})
		}
	}
	receipt.FirstUnresolved = nil
	if len(receipt.UnresolvedClaims) > 0 {
		receipt.FirstUnresolved = &receipt.UnresolvedClaims[0]
	}
	if runtime.Observation.Failure != "" {
		receipt.Decision = "FAIL_CLOSED"
		receipt.FailClosedReason = &runtime.Observation.Failure
	}
}

func replaceDiscoveryDimension(receipt *completeness.CompletenessReceipt, dimension completeness.CompletenessDimension) {
	for i, existing := range receipt.Dimensions {
		if existing.ID == dimension.ID {
			receipt.Dimensions[i] = dimension
			return
		}
	}
	receipt.Dimensions = append(receipt.Dimensions, dimension)
	receipt.CoreDimensions = append(receipt.CoreDimensions, dimension.ID)
}
