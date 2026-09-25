package languagegointeroperation

import (
	"runtime"
	"strings"
)

func Evaluate(input Input) Report {
	registry := Registry()
	registryDigest := digestJSON(registry)
	source := newSource(input, registryDigest, runtime.Version())
	if input.ExpectedHeadSHA == "" {
		return failureReport(source, "EXPECTED_HEAD_UNKNOWN")
	}
	if !strings.HasPrefix(runtime.Version(), RequiredGoVersion) {
		return failureReport(source, "GO_1_27_TOOLCHAIN_REQUIRED")
	}
	if err := registry.Validate(); err != nil {
		return failureReport(source, "INTEROP_REGISTRY_INVALID")
	}
	observation, err := observeConcept(input.ConceptArtifact)
	if err != nil {
		return failureReport(source, "CONCEPT_EVIDENCE_UNKNOWN")
	}
	source.ConceptBound = observation.Drift == 0
	results := executeCases(registry.Cases)
	summary := summarize(registry.Cases, results, observation.Drift)
	return successOrFailureReport(source, summary, results)
}

func newSource(input Input, registryDigest, toolchain string) Source {
	return Source{ExpectedHeadSHA: input.ExpectedHeadSHA, ConceptID: ConceptID,
		Producer: "languagegointeroperation.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: ExpectedMetaOperation, ConceptArtifactDigest: input.ConceptArtifact.ArtifactDigest,
		CatalogDigest: input.ConceptArtifact.CatalogDigest, RegistryDigest: registryDigest,
		Toolchain: toolchain, GoReleaseNotes: "https://go.dev/doc/go1.27",
		MacroReference: "https://github.com/cosmos72/gomacro"}
}
