package main

import (
	"os"
	"path/filepath"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func observeAdmissionCase(cfg config, f fixture, baseline variant, planned countercase) (admissionCase, error) {
	row := admissionCase{ID: planned.ID, ExpectedReason: planned.Reason}
	v, err := observedAdmissionVariant(baseline, f, planned)
	if err != nil {
		return row, err
	}
	row.InputDigests = inputDigests(v)
	if err := saveVariant(filepath.Join(cfg.Output, "observed-admission", planned.ID), cfg.Origin, f, v); err != nil {
		return row, err
	}
	v.Inputs.ScratchDirectory = filepath.Join(cfg.Scratch, "observed-admission-source", planned.ID)
	_, sourceErr := co.VerifyEvidence(v.Receipt, v.Inputs, os.DirFS(cfg.Root), cfg.Root,
		cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if sourceErr != nil {
		row.SourceReason = sourceErr.Error()
	}
	first, firstErr := readiness.BuildWithConceptOperationEvidence(f.Concept, v.Receipt, v.Inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-admission-first", planned.ID),
		cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	replay, replayErr := readiness.BuildWithConceptOperationEvidence(f.Concept, v.Receipt, v.Inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-admission-replay", planned.ID),
		cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	row.ResultProduced = firstErr == nil || replayErr == nil
	row.ReadinessCredit = admissionCredit(first) || admissionCredit(replay)
	if firstErr != nil {
		row.FirstConsumerReason = firstErr.Error()
	}
	if replayErr != nil {
		row.ReplayConsumerReason = replayErr.Error()
	}
	row.Passed = sourceErr != nil && firstErr != nil && replayErr != nil &&
		row.SourceReason == planned.Reason && row.FirstConsumerReason == planned.Reason &&
		row.ReplayConsumerReason == planned.Reason && !row.ResultProduced && !row.ReadinessCredit
	return row, nil
}

func admissionCredit(receipt readiness.Receipt) bool {
	for _, obligation := range receipt.Snapshot.Obligations {
		if obligation.ID == "META-CONCEPT-GOVERNED-REFACTORING" && obligation.Status == "SATISFIED" {
			return true
		}
	}
	return false
}
