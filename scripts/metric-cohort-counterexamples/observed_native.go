package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func observedNativeCase(baseline co.Receipt, plan observedCasePlan) bodyexecution.CompositionCase {
	expected := make([]string, 0, baseline.ExpectedCount)
	for _, operation := range baseline.Expected {
		expected = append(expected, operation.Operation)
	}
	declared, _ := json.Marshal(map[string]any{"ids": expected})
	observed, _ := json.Marshal(map[string]any{"ids": plan.Expected.ObservedIDs})
	wanted, _ := json.Marshal(plan.Expected)
	return bodyexecution.CompositionCase{
		Inputs:   map[string]json.RawMessage{"AssessDeclared.input0": declared, "AssessDeclared.input1": observed},
		Expected: map[string]json.RawMessage{"AssessDeclared": wanted},
	}
}

func executeObservedNative(cfg config, source []byte,
	suite bodyexecution.CompositionCases) (bodyexecution.CompositionRuntime, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	options := bodyexecution.CompositionOptions{EntryActivity: co.ObservationPolicyActivity}
	prior, err := bodyexecution.GenerateCompositionWithOptions(ctx, co.ObservationPolicyPath, source, suite, options)
	if err != nil {
		return bodyexecution.CompositionRuntime{}, err
	}
	tool := "go"
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	native, executionErr := bodyexecution.ExecuteComposition(ctx, co.ObservationPolicyPath, source,
		prior, suite, filepath.Join(runtime.GOROOT(), "bin", tool))
	raw, err := json.MarshalIndent(native, "", "  ")
	if err != nil {
		return native, err
	}
	if err := os.WriteFile(filepath.Join(cfg.Output, "observed-native.json"), append(raw, '\n'), 0640); err != nil {
		return native, err
	}
	return native, executionErr
}
