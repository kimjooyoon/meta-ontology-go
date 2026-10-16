package main

import (
	"fmt"
	"os"
	"path/filepath"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func consumeObservedCase(cfg config, f fixture, inputs co.SourceInputs, payload []byte,
	row observedCaseResult, baseline readiness.Receipt) (observedCaseResult, error) {
	inputs.ScratchDirectory = filepath.Join(cfg.Scratch, "observed-source", row.ID)
	state, err := co.VerifyEvidence(payload, inputs, os.DirFS(cfg.Root), cfg.Root,
		cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	row.SourceReplayed = err == nil && state.Status == row.Expected.Status && state.Digest == row.Receipt.Digest
	if !row.SourceReplayed {
		return row, fmt.Errorf("separate observation source replay failed: %v", err)
	}
	first, firstErr := readiness.BuildWithConceptOperationEvidence(f.Concept, payload, inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-first", row.ID), cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	replay, replayErr := readiness.BuildWithConceptOperationEvidence(f.Concept, payload, inputs,
		cfg.Root, filepath.Join(cfg.Scratch, "observed-replay", row.ID), cfg.Origin.Repository, cfg.Origin.SubjectSHA)
	if firstErr != nil {
		row.FirstConsumerReason = firstErr.Error()
	}
	if replayErr != nil {
		row.ReplayConsumerReason = replayErr.Error()
	}
	if row.Expected.Status != "VERIFIED" {
		row.Passed = firstErr != nil && replayErr != nil &&
			row.FirstConsumerReason == "concept-operation evidence is not verified" &&
			row.FirstConsumerReason == row.ReplayConsumerReason
		return row, nil
	}
	if firstErr != nil || replayErr != nil || !artifact.Equal(first, replay) || verifyPositive(first) != nil {
		return row, fmt.Errorf("verified observation readiness did not replay")
	}
	row.ReadinessCredit = true
	intact := sameUnrelatedObligations1889(baseline, first)
	row.UnrelatedObligationsIntact, row.Passed = &intact, intact
	return row, nil
}

func sameUnrelatedObligations1889(baseline, observed readiness.Receipt) bool {
	if len(baseline.Snapshot.Obligations) != 24 || len(observed.Snapshot.Obligations) != 24 {
		return false
	}
	matched := 0
	for index, before := range baseline.Snapshot.Obligations {
		after := observed.Snapshot.Obligations[index]
		if before.ID == "META-CONCEPT-GOVERNED-REFACTORING" {
			continue
		}
		if !artifact.Equal(before, after) {
			return false
		}
		matched++
	}
	return matched == 23
}
