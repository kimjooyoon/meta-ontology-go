package analyzer

import (
	"go/token"
)

// Analyzer extracts semantic deltas using a fixed registry.
type Analyzer struct {
	registry *Registry
}

// New creates an Analyzer. A nil registry behaves as an empty registry.
func New(registry *Registry) *Analyzer {
	return &Analyzer{registry: registry}
}

// AnalyzeSource analyzes one Go source file. PackagePath is inferred from the
// package declaration, so callers needing cross-module disambiguation should
// use AnalyzePackage with an explicit SourceFile.PackagePath.
func AnalyzeSource(filename string, source []byte, registry *Registry) (Result, error) {
	return New(registry).AnalyzePackage([]SourceFile{{Filename: filename, Source: source}})
}

// Analyze is a short alias for AnalyzeSource.
func Analyze(filename string, source []byte, registry *Registry) (Result, error) {
	return AnalyzeSource(filename, source, registry)
}

// AnalyzePackage analyzes one or more files in one Go package. Annotations
// are collected before references are visited, and all output is sorted.
func (a *Analyzer) AnalyzePackage(sources []SourceFile) (Result, error) {
	if len(sources) == 0 {
		return Result{}, nil
	}

	fileSet := token.NewFileSet()
	parsed, err := parseSources(fileSet, sources)
	if err != nil {
		return Result{}, err
	}
	registry := (*Registry)(nil)
	if a != nil {
		registry = a.registry
	}
	resolver := newResolver(registry, fileSet, parsed)
	result := Result{}
	for _, file := range parsed {
		registrations, diagnostics := collectRegistrations(file, fileSet)
		result.Diagnostics = append(result.Diagnostics, diagnostics...)
		for _, registration := range registrations {
			if resolver.addLocal(registration) {
				result.Registrations = append(result.Registrations, registration)
			}
		}
	}
	for _, file := range parsed {
		resolver.analyzeFile(file, &result.Delta)
	}
	result.Delta.sort()
	sortRegistrations(result.Registrations)
	result.Diagnostics = result.Diagnostics.SortBySpan()
	return result, nil
}

// AnalyzeFile analyzes one Go source file with this Analyzer's registry.
func (a *Analyzer) AnalyzeFile(filename string, source []byte) (Result, error) {
	return a.AnalyzePackage([]SourceFile{{Filename: filename, Source: source}})
}

// AnalyzePackage is the package-level convenience form of Analyzer.AnalyzePackage.
func AnalyzePackage(sources []SourceFile, registry *Registry) (Result, error) {
	return New(registry).AnalyzePackage(sources)
}
