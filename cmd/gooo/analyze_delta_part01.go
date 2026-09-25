package main

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func readAnalyzeAuthority(filename string, reader SourceReader, parser SourceParser, deadline time.Time) (semantic.IR, generator.SemanticIR, error) {
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		return semantic.IR{}, generator.SemanticIR{}, err
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		return semantic.IR{}, generator.SemanticIR{}, err
	}
	if diagnostics.HasErrors() {
		return semantic.IR{}, generator.SemanticIR{}, diagnostics.Error()
	}
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), bidir.Lower)
	if err != nil {
		return semantic.IR{}, generator.SemanticIR{}, err
	}
	model, err := projectionIR(ir)
	return ir, model, err
}
func readAnalyzeSources(files []string, reader SourceReader, model generator.SemanticIR, authority semantic.IR, deadline time.Time) ([]analyzer.SourceFile, error) {
	sources := make([]analyzer.SourceFile, 0, len(files))
	for _, filename := range files {
		source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filename, err)
		}
		if err := validateAnalyzeGeneratedSource(model, authority, source); err != nil {
			return nil, fmt.Errorf("%s: %w", filename, err)
		}
		logicalFilename := filename
		if len(files) == 1 {
			logicalFilename = filepath.Base(filename)
		}
		sources = append(sources, analyzer.SourceFile{Filename: logicalFilename, PackagePath: authority.Package, Source: source})
	}
	return sources, nil
}
func reportAnalyzeDeltaError(stderr io.Writer, filename, phase string, err error) int {
	if filename != "" {
		fmt.Fprintf(stderr, "gooo: %s: analyze: %s: %v\n", filename, phase, err)
	} else {
		fmt.Fprintf(stderr, "gooo: analyze: %s: %v\n", phase, err)
	}
	return exitFailure
}
func analyzeMappingPolicy() (analyzer.MappingPolicy, error) {
	p, err := analyzer.NewMappingPolicy(analyzer.CurrentSemanticAdapterPolicy)
	if err != nil {
		return analyzer.MappingPolicy{}, err
	}
	for _, m := range []analyzer.RelationMapping{
		{Source: analyzer.RelationUses, Predicate: semantic.Used, SourceSubjectKind: semantic.Activity, SourceObjectKind: semantic.Entity, AllowedOrigins: []analyzer.ObservationOrigin{analyzer.OriginSignature}},
		{Source: analyzer.RelationGenerates, Predicate: semantic.WasGeneratedBy, SourceSubjectKind: semantic.Activity, SourceObjectKind: semantic.Entity, Reverse: true, AllowedOrigins: []analyzer.ObservationOrigin{analyzer.OriginSignature}},
	} {
		if err := p.Register(m); err != nil {
			return analyzer.MappingPolicy{}, err
		}
	}
	return p, nil
}
