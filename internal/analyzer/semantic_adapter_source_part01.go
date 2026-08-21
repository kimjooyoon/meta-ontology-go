package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

const sourceBundleSchema = "analyzer-source-bundle/v1"

// SourceSemanticAdapterInput binds raw Go source to one typed semantic
// adaptation. The source bytes are hashed before parsing so the result cannot
// silently claim evidence for a different source presentation.
type SourceSemanticAdapterInput struct {
	Base              semantic.IR
	Sources           []SourceFile
	Registry          *Registry
	Policy            MappingPolicy
	Producer          semantic.ID
	EvidenceKind      semantic.EvidenceKind
	ToolchainIdentity string
}

// AnalyzeAndAdaptSemantic performs the registered Go analysis and typed
// semantic adaptation as one source-bound operation. It does not write files
// or mutate the supplied source slices or base IR.
func AnalyzeAndAdaptSemantic(input SourceSemanticAdapterInput) (SemanticAdapterResult, error) {
	if err := input.Policy.Validate(); err != nil {
		return SemanticAdapterResult{}, err
	}
	toolchain := strings.TrimSpace(input.ToolchainIdentity)
	if toolchain == "" {
		return SemanticAdapterResult{}, adapterError(AdapterSourceConfig, "", "", "toolchain identity is required")
	}
	sources, err := canonicalSourceFiles(input.Sources)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	sourceDigest, err := SourceBundleDigest(sources)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	slots, err := collectProtectedSlots(sources)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	base, err := input.Base.Normalized()
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	boundSlots := bindProtectedSlots(
		slots, sourceDigest, base.StableHash(), input.Policy.Digest(), ToolchainDigest(toolchain),
		input.Registry.Digest(),
	)
	analysis, err := AnalyzePackage(sources, input.Registry)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	if analysis.Diagnostics.HasErrors() {
		return SemanticAdapterResult{}, adapterError(
			AdapterAnalysisDiagnostics, "", "", analysis.Diagnostics.Error().Error(),
		)
	}
	analysis.Registrations = append(analysis.Registrations, input.Registry.all()...)
	sortRegistrations(analysis.Registrations)
	return AdaptSemantic(SemanticAdapterInput{
		Base: input.Base, Analysis: analysis, Policy: input.Policy,
		Producer: input.Producer, EvidenceKind: input.EvidenceKind,
		SourceDigest: sourceDigest, ToolchainDigest: ToolchainDigest(toolchain), Registry: input.Registry,
		SlotObservations: boundSlots,
	})
}
