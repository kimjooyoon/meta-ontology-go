package main

import (
	"encoding/json"
	"path/filepath"
	"time"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func executeCase(cfg config, f fixture, planned countercase) (result caseResult, executionErr error) {
	started := time.Now()
	defer func() { result.ElapsedNanoseconds = time.Since(started).Nanoseconds() }()
	row := caseResult{ID: planned.ID, Kind: planned.Kind, ExpectedReason: planned.Reason}
	v, err := buildVariant(f, planned)
	if err != nil {
		row.HarnessError = err.Error()
		return row, err
	}
	row.ProducerError, row.ProducerReceipt = v.ProducerError, v.Producer
	row.ProducerAttempted = v.ProducerAttempted
	if planned.ID == "ambiguous-binding" {
		row.ExpectedProducerError = expectedDuplicateReason(f)
	}
	row.InputDigests = inputDigests(v)
	if err := saveVariant(filepath.Join(cfg.Output, "cases", planned.ID), cfg.Origin, f, v); err != nil {
		row.HarnessError = err.Error()
		return row, err
	}
	_, envelopeErr := co.Verify(v.Receipt, cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	row.EnvelopeAccepted = envelopeErr == nil
	if len(v.Receipt) != 0 {
		var decoded co.Receipt
		if err := json.Unmarshal(v.Receipt, &decoded); err == nil {
			row.ClaimedExpectedCount, row.ClaimedCoverageBPS = &decoded.ExpectedCount, &decoded.CoverageBPS
		}
	}
	scratch := filepath.Join(cfg.Scratch, planned.ID)
	row.ConsumerAttempted = true
	observed, consumerErr := readiness.BuildWithConceptOperationEvidence(f.Concept, v.Receipt, v.Inputs,
		cfg.Root, scratch, cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	row.ReadinessResultProduced = consumerErr == nil
	if consumerErr != nil {
		row.ObservedReason = consumerErr.Error()
	} else {
		for _, obligation := range observed.Snapshot.Obligations {
			if obligation.ID == "META-CONCEPT-GOVERNED-REFACTORING" && obligation.Status == "SATISFIED" {
				row.ReadinessCredit = true
			}
		}
	}
	row.Passed = consumerErr != nil && row.ObservedReason == planned.Reason &&
		(!planned.EnvelopeRequired || row.EnvelopeAccepted) && producerObservation(planned, v, f)
	return row, nil
}
