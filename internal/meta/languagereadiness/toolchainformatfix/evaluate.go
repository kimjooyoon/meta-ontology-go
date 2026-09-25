package toolchainformatfix

import (
	"runtime"
	"strings"
)

func Evaluate(input Input) Report {
	source := Source{ExpectedHeadSHA: input.ExpectedHeadSHA, GoVersion: runtime.Version(),
		ConceptDigest: input.ConceptArtifact.ArtifactDigest, RegistryDigest: digestBytes(input.RegistryRaw)}
	registry, registryErr := DecodeRegistry(input.RegistryRaw)
	if registryErr == nil {
		source.RegistryDigest = digestJSON(registry)
	}
	if input.Executor != nil {
		source.ExecutableDigest, _ = input.Executor.BinaryDigest()
	}
	source.ObservationKnown = registryErr == nil && validHead(input.ExpectedHeadSHA) &&
		strings.HasPrefix(runtime.Version(), "go1.27") && input.ConceptArtifact.Ready() &&
		input.ConceptArtifact.Report.Summary.Concepts >= 21 && validDigest(source.ExecutableDigest)
	report := Report{Schema: ReportSchema, Source: source}
	if !source.ObservationKnown || input.Executor == nil {
		for _, definition := range expectedRegistry().Cases {
			arguments, _ := argumentsFor(definition.Operation)
			report.Cases = append(report.Cases, unresolvedCase(
				CaseResult{Definition: definition, Arguments: arguments},
				"FORMAT_FIX_EVIDENCE_UNKNOWN"))
		}
		return finish(report, cycleEvidence{}, 1, "FORMAT_FIX_EVIDENCE_UNKNOWN")
	}
	for _, definition := range registry.Cases {
		report.Cases = append(report.Cases, executeCase(input.Executor, definition))
	}
	return finish(report, evaluateCycle(), 0, "FORMAT_FIX_OBSERVATION_UNKNOWN")
}
