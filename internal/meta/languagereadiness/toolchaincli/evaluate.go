package toolchaincli

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
		input.ConceptArtifact.Report.Summary.Concepts >= 20 && validDigest(source.ExecutableDigest)
	report := Report{Schema: ReportSchema, Source: source}
	if !source.ObservationKnown || input.Executor == nil {
		return lowerResolution(report, "TOOLCHAIN_CLI_EVIDENCE_UNKNOWN")
	}
	for _, definition := range registry.Cases {
		report.Cases = append(report.Cases, executeCase(input.Executor, definition))
	}
	return finish(report, 0, "TOOLCHAIN_CLI_OBSERVATION_UNKNOWN")
}

func lowerResolution(report Report, reason string) Report {
	for _, definition := range expectedRegistry().Cases {
		arguments, _ := argumentsFor(definition.Operation)
		result := CaseResult{Definition: definition, Arguments: arguments}
		report.Cases = append(report.Cases, unresolvedCase(result))
	}
	return finish(report, 1, reason)
}
