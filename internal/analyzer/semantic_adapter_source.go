package analyzer

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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
	)
	analysis, err := AnalyzePackage(sources, input.Registry)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	analysis.Registrations = append(analysis.Registrations, input.Registry.all()...)
	sortRegistrations(analysis.Registrations)
	return AdaptSemantic(SemanticAdapterInput{
		Base: input.Base, Analysis: analysis, Policy: input.Policy,
		Producer: input.Producer, EvidenceKind: input.EvidenceKind,
		SourceDigest: sourceDigest, ToolchainDigest: ToolchainDigest(toolchain),
		SlotObservations: boundSlots,
	})
}

// SourceBundleDigest returns the schema-bound SHA-256 identity of raw source
// files. File order is presentation; filename, package path, and exact bytes
// are identity. Duplicate filenames are rejected rather than merged.
func SourceBundleDigest(sources []SourceFile) (string, error) {
	canonical, err := canonicalSourceFiles(sources)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(sourceBundleSchema)
	b.WriteByte('\n')
	for _, source := range canonical {
		writeSourceField(&b, source.Filename)
		writeSourceField(&b, source.PackagePath)
		b.WriteString(strconv.Itoa(len(source.Source)))
		b.WriteByte(':')
		b.Write(source.Source)
		b.WriteByte('\n')
	}
	return semantic.StableHashString(b.String()), nil
}

func canonicalSourceFiles(sources []SourceFile) ([]SourceFile, error) {
	if len(sources) == 0 {
		return nil, adapterError(AdapterSourceConfig, "", "", "at least one source file is required")
	}
	copyOf := make([]SourceFile, len(sources))
	seen := make(map[string]struct{}, len(sources))
	for index, source := range sources {
		filename := source.Filename
		if filename == "" {
			filename = "<source>"
		}
		if _, exists := seen[filename]; exists {
			return nil, adapterError(AdapterSourceConfig, "", filename, "duplicate source filename")
		}
		seen[filename] = struct{}{}
		copyOf[index] = SourceFile{
			Filename: filename, PackagePath: source.PackagePath,
			Source: append([]byte(nil), source.Source...),
		}
	}
	sort.Slice(copyOf, func(i, j int) bool {
		if copyOf[i].Filename != copyOf[j].Filename {
			return copyOf[i].Filename < copyOf[j].Filename
		}
		return copyOf[i].PackagePath < copyOf[j].PackagePath
	})
	return copyOf, nil
}

func writeSourceField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('\n')
}
