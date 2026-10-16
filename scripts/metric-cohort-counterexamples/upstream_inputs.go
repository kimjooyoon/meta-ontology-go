package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
	strategyverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/verify"
)

const unregisteredUpstreamOperation = "gooo://native/unregistered-upstream/operation-1743"

// This adds an actual upstream identity before strategy synthesis. Resealed
// adversarial input remains diagnostic input, never upstream execution proof.
func extendUnregisteredLedger(inputs co.SourceInputs) (co.SourceInputs, []metricstrategy.ConceptOperationSpec, error) {
	var ledger metric.Ledger
	if err := json.Unmarshal(inputs.Intervention, &ledger); err != nil {
		return inputs, nil, err
	}
	for _, indicator := range ledger.Indicators {
		if indicator.MetaOperation == unregisteredUpstreamOperation {
			return inputs, nil, fmt.Errorf("unregistered upstream case already exists in baseline")
		}
	}
	ledger.Indicators = append(ledger.Indicators, metric.Indicator{
		ID: "MIV-FOUNDATION-UPSTREAM-UNREGISTERED-1743", Family: "FOUNDATION", Trilemma: "AXIOM",
		MetaOperation: unregisteredUpstreamOperation, Expected: "ADVERSARIAL_INPUT", Actual: "ADVERSARIAL_INPUT",
		Status: "SATISFIED", EvidenceDigest: artifact.ContentDigest([]byte("unregistered-upstream-input-1743")),
	})
	var err error
	ledger, err = metric.SealLedger(ledger)
	if err != nil {
		return inputs, nil, err
	}
	cohort, err := metricstrategy.ConceptOperationCohort(ledger.Indicators)
	if err != nil {
		return inputs, nil, err
	}
	var verification metricverify.Receipt
	if err := json.Unmarshal(inputs.InterventionVerification, &verification); err != nil {
		return inputs, nil, err
	}
	verification.LedgerDigest, verification.IndicatorCount = ledger.Digest, len(ledger.Indicators)
	verification.Digest = ""
	verification.Digest, err = artifact.Digest(verification)
	if err != nil {
		return inputs, nil, err
	}
	inputs.Intervention, err = json.Marshal(ledger)
	if err == nil {
		inputs.InterventionVerification, err = json.Marshal(verification)
	}
	inputs.Program, inputs.ProgramSource, inputs.ProgramVerification = nil, nil, nil
	return inputs, cohort, err
}

func synthesizeUnregisteredStrategy(cfg config, inputs co.SourceInputs) (co.SourceInputs, error) {
	directory := filepath.Join(cfg.Scratch, "unregistered-upstream-inputs")
	if err := os.Mkdir(directory, 0750); err != nil {
		return inputs, err
	}
	for name, raw := range map[string][]byte{
		"source-metrics.json": inputs.SourceMetrics, "intervention-ledger.json": inputs.Intervention,
		"intervention-verification.json": inputs.InterventionVerification,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), raw, 0640); err != nil {
			return inputs, err
		}
	}
	metrics, ledger, verification := filepath.Join(directory, "source-metrics.json"),
		filepath.Join(directory, "intervention-ledger.json"), filepath.Join(directory, "intervention-verification.json")
	plan, err := metricstrategy.Generate(os.DirFS(cfg.Root), metrics, ledger, verification,
		cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if err != nil {
		return inputs, err
	}
	replay, err := strategyverify.Replay(os.DirFS(cfg.Root), metrics, ledger, verification, plan)
	if err != nil {
		return inputs, err
	}
	inputs.Strategy, err = json.Marshal(plan)
	if err == nil {
		inputs.StrategyVerification, err = json.Marshal(replay)
	}
	return inputs, err
}
