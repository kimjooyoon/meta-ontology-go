package externalecosystemexecution

import (
	"context"
	"path/filepath"
	"sort"
)

const referenceEvidencePath = "internal/meta/externalecosystemconformance/evidence/gomacro.json"

func Observe(ctx context.Context, sourceRoot, externalRoot string) (Observation, error) {
	sourceBefore, err := captureRepository(ctx, sourceRoot)
	if err != nil {
		return Observation{}, err
	}
	externalBefore, err := captureRepository(ctx, externalRoot)
	if err != nil {
		return Observation{}, err
	}
	goVersion, err := commandText(ctx, sourceRoot, "go", "env", "GOVERSION")
	if err != nil {
		return Observation{}, err
	}
	moduleGo, err := moduleGoVersion(filepath.Join(externalRoot, "go.mod"))
	if err != nil {
		return Observation{}, err
	}
	runs := []RunObservation{runGoTest(ctx, externalRoot, 1), runGoTest(ctx, externalRoot, 2)}
	externalAfter, err := captureRepository(ctx, externalRoot)
	if err != nil {
		return Observation{}, err
	}
	sourceAfter, err := captureRepository(ctx, sourceRoot)
	if err != nil {
		return Observation{}, err
	}
	return Observation{
		Schema: ObservationSchema, Reference: loadReference(sourceRoot, externalBefore, moduleGo),
		GoVersion: goVersion, Runs: runs, SourceBefore: sourceBefore, SourceAfter: sourceAfter,
		ExternalBefore: externalBefore, ExternalAfter: externalAfter,
	}, nil
}

func diagnosticEvent(event goEvent) bool {
	if event.Action == "build-output" {
		return event.Output != ""
	}
	return event.Action == "output" && event.Output != "" &&
		(event.OutputType == "error" || event.OutputType == "error-continue")
}

func normalizedEventResults(final map[string]Outcome, unknown map[string]bool) ([]Outcome, []string) {
	outcomes := make([]Outcome, 0, len(final))
	for _, item := range final {
		outcomes = append(outcomes, item)
	}
	sort.Slice(outcomes, func(i, j int) bool {
		if outcomes[i].Package == outcomes[j].Package {
			return outcomes[i].Test < outcomes[j].Test
		}
		return outcomes[i].Package < outcomes[j].Package
	})
	unknowns := make([]string, 0, len(unknown))
	for item := range unknown {
		unknowns = append(unknowns, item)
	}
	sort.Strings(unknowns)
	return outcomes, unknowns
}
