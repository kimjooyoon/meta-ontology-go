package languagediagnosticprovenance

import (
	"reflect"
	"runtime"
	"strings"
)

func Evaluate(input Input) Report {
	registryDigest := digestJSON(input.Registry)
	source := newSource(input, registryDigest, runtime.Version())
	if input.ExpectedHeadSHA == "" {
		return failureReport(source, "EXPECTED_HEAD_UNKNOWN")
	}
	if !strings.HasPrefix(runtime.Version(), RequiredGoVersion) {
		return failureReport(source, "GO_1_27_TOOLCHAIN_REQUIRED")
	}
	if err := input.Registry.Validate(); err != nil {
		return failureReport(source, "DIAGNOSTIC_PROVENANCE_REGISTRY_INVALID")
	}
	if !reflect.DeepEqual(input.Registry, Registry()) {
		return failureReport(source, "DIAGNOSTIC_PROVENANCE_REGISTRY_DRIFT")
	}
	conceptDrift, err := observeConcept(input.ConceptArtifact)
	if err != nil {
		return failureReport(source, "CONCEPT_EVIDENCE_UNKNOWN")
	}
	results := executeCases(input.Registry.Cases)
	summary := summarize(input.Registry.Cases, results, conceptDrift)
	return successOrFailureReport(source, summary, results)
}

func newSource(input Input, registryDigest, toolchain string) Source {
	return Source{
		ExpectedHeadSHA: input.ExpectedHeadSHA, ConceptID: ConceptID,
		Producer: "languagediagnosticprovenance.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: ExpectedMetaOperation,
		ConceptArtifactDigest: input.ConceptArtifact.ArtifactDigest,
		CatalogDigest: input.ConceptArtifact.CatalogDigest,
		RegistryDigest: registryDigest, Toolchain: toolchain,
		TokenReference: "https://pkg.go.dev/go/token",
		ScannerReference: "https://pkg.go.dev/go/scanner",
		TypesReference: "https://pkg.go.dev/go/types",
		LineDirective: "https://go.dev/wiki/Comments#line-directives",
		MacroReference: "https://github.com/cosmos72/gomacro",
	}
}
