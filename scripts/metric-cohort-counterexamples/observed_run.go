package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func observeSeparateObservations(cfg config, f fixture) (observedCohortReport, error) {
	result := observedCohortReport{Decision: "FAIL_CLOSED"}
	if f.Receipt.ExpectedCount == 0 {
		return result, fmt.Errorf("observed cohort has no declared source contract")
	}
	policy, err := os.ReadFile(filepath.Join(cfg.Root, co.ObservationPolicyPath))
	if err != nil {
		return result, err
	}
	baseline, err := readiness.BuildWithConceptOperationEvidence(f.Concept, f.Raw, f.Inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-baseline"), cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if err != nil {
		return result, err
	}
	plans := observedPlans1889(f.Receipt.ExpectedCount)
	result.CasesDeclared = len(plans)
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, plan := range plans {
		row, nativeCase, err := produceObservedCase(cfg, f, policy, plan, baseline)
		result.Cases = append(result.Cases, row)
		if err != nil {
			result.SetupError = err.Error()
			return result, err
		}
		suite.Cases = append(suite.Cases, nativeCase)
		if row.Passed {
			result.CasesPassed++
		}
	}
	native, err := executeObservedNative(cfg, policy, suite)
	result.NativeRuns, result.NativePassed, result.ModelCalls = len(native.Runs), native.FinitePassed, native.ModelCalls
	if err != nil {
		result.SetupError = err.Error()
		return result, err
	}
	if result.CasesPassed == result.CasesDeclared && native.RuntimeReplayed &&
		native.FinitePassed == result.CasesDeclared && native.FiniteTotal == result.CasesDeclared &&
		len(native.Runs) == 2 && native.ModelCalls == 0 {
		result.Decision = "PASS"
		return result, nil
	}
	return result, fmt.Errorf("separate observation/native expectations did not all pass")
}

func produceObservedCase(cfg config, f fixture, policy []byte, plan observedCasePlan,
	baseline readiness.Receipt) (observedCaseResult, bodyexecution.CompositionCase, error) {
	row := observedCaseResult{ID: plan.ID}
	document, err := co.DeclareObservations(f.Receipt, f.Inputs.Strategy)
	if err != nil {
		return row, bodyexecution.CompositionCase{}, err
	}
	reviseObserved1889(&document, plan.ID)
	plan.Expected.ObservedIDs = make([]string, 0, len(document.Observations))
	for _, observed := range document.Observations {
		plan.Expected.ObservedIDs = append(plan.Expected.ObservedIDs, observed.Operation)
	}
	row.Expected = plan.Expected
	inputs := f.Inputs
	raw, err := json.Marshal(document)
	if err != nil {
		return row, bodyexecution.CompositionCase{}, err
	}
	inputs.Observation = &co.ObservationInputs{Document: raw, PolicySource: policy}
	receipt, err := co.BuildObserved(context.Background(), inputs)
	if err != nil {
		return row, bodyexecution.CompositionCase{}, err
	}
	row.Receipt = &receipt
	var actual co.ObservationAssessment
	if json.Unmarshal(receipt.Decision.Value, &actual) != nil || !artifact.Equal(actual, plan.Expected) ||
		receipt.Baseline.Digest != f.Receipt.Digest {
		return row, bodyexecution.CompositionCase{}, fmt.Errorf("source observation differs from independent expectation: %s", plan.ID)
	}
	payload, err := json.Marshal(receipt)
	if err != nil {
		return row, bodyexecution.CompositionCase{}, err
	}
	if err := saveVariant(filepath.Join(cfg.Output, "observed-cases", plan.ID), cfg.Origin, f,
		variant{Receipt: payload, Inputs: inputs, ProducerAttempted: true}); err != nil {
		return row, bodyexecution.CompositionCase{}, err
	}
	row, err = consumeObservedCase(cfg, f, inputs, payload, row, baseline)
	return row, observedNativeCase(f.Receipt, plan), err
}
