package main

import (
	"encoding/json"
	"fmt"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type producerAttempt struct {
	Repository string `json:"repository"`
	SubjectSHA string `json:"subject_sha"`
	RunID      int64  `json:"workflow_run_id"`
	Attempt    int    `json:"workflow_run_attempt"`
	Job        string `json:"producer_job"`
}

func loadFixture(cfg config) (fixture, error) {
	files, err := transportPayloads(cfg)
	if err != nil {
		return fixture{}, err
	}
	var attempt producerAttempt
	if err := json.Unmarshal(files["producer-attempt.json"], &attempt); err != nil {
		return fixture{}, err
	}
	if attempt.Repository != cfg.Origin.Repository || attempt.SubjectSHA != cfg.Origin.SubjectSHA ||
		attempt.RunID != cfg.Origin.RunID || attempt.Attempt != cfg.Origin.Attempt || attempt.Job != "program" {
		return fixture{}, fmt.Errorf("producer payload differs from original run identity")
	}
	f := fixture{Raw: files["concept-operation-binding.json"], Inputs: co.SourceInputs{
		Strategy: files["strategy-plan.json"], StrategyVerification: files["strategy-verification.json"],
		SourceMetrics: files["source-metrics.json"], Intervention: files["intervention-ledger.json"],
		InterventionVerification: files["intervention-verification.json"], Program: files["program.json"],
		ProgramSource: files["program.gooo"], ProgramVerification: files["verification.json"],
	}}
	f.Receipt, err = co.Verify(f.Raw, cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if err != nil {
		return f, err
	}
	var ledger metric.Ledger
	if err := json.Unmarshal(f.Inputs.Intervention, &ledger); err != nil {
		return f, err
	}
	f.Contract, err = metricstrategy.ConceptOperationCohort(ledger.Indicators)
	if err != nil {
		return f, err
	}
	f.Concept, err = readBounded(cfg.Concept, 32<<20)
	return f, err
}
