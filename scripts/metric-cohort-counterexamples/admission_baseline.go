package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func observedAdmissionBaseline(cfg config, f fixture) (variant, co.ObservedReceipt, error) {
	inputs := f.Inputs
	policy, err := os.ReadFile(filepath.Join(cfg.Root, co.ObservationPolicyPath))
	if err != nil {
		return variant{}, co.ObservedReceipt{}, err
	}
	document, err := co.DeclareObservations(f.Receipt, inputs.Strategy)
	if err != nil {
		return variant{}, co.ObservedReceipt{}, err
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return variant{}, co.ObservedReceipt{}, err
	}
	inputs.Observation = &co.ObservationInputs{Document: raw, PolicySource: policy}
	receipt, err := co.BuildObserved(context.Background(), inputs)
	if err != nil {
		return variant{}, receipt, err
	}
	if receipt.Baseline.Digest != f.Receipt.Digest {
		return variant{}, receipt, fmt.Errorf("observed admission baseline changed the original receipt")
	}
	payload, err := json.Marshal(receipt)
	if err != nil {
		return variant{}, receipt, err
	}
	observed, err := readiness.BuildWithConceptOperationEvidence(f.Concept, payload, inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-admission-baseline"), cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if err != nil {
		return variant{}, receipt, err
	}
	if err := verifyPositive(observed); err != nil {
		return variant{}, receipt, err
	}
	v := variant{Receipt: payload, Inputs: inputs}
	err = saveVariant(filepath.Join(cfg.Output, "observed-admission", "baseline"), cfg.Origin, f, v)
	return v, receipt, err
}
