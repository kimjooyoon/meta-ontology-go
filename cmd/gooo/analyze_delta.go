package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
		sources = append(sources, analyzer.SourceFile{Filename: filename, PackagePath: authority.Package, Source: source})
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

func analyzeRegistry(authority semantic.IR, sources []analyzer.SourceFile) (*analyzer.Registry, error) {
	r := analyzer.NewRegistry()
	byID, byName := map[string]semantic.Node{}, map[string][]semantic.Node{}
	for _, node := range authority.Graph.Nodes() {
		if node.Kind != semantic.Entity && node.Kind != semantic.Activity {
			continue
		}
		byID[string(node.ID)] = node
		byName[node.Name] = append(byName[node.Name], node)
		if err := r.Register(analyzer.Registration{Ref: analyzer.SymbolRef{PackagePath: authority.Package, PackageName: authority.Package, Name: node.Name}, Kind: analyzeSymbolKind(node.Kind), Identity: analyzer.NewIdentity(node.Namespace.String(), string(node.ID))}); err != nil {
			return nil, err
		}
	}
	for _, source := range sources {
		aliases, err := analyzeMarkerAliases(source.Source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source.Filename, err)
		}
		for _, alias := range aliases {
			node, ok := byID[alias.id]
			if !ok {
				return nil, fmt.Errorf("generated marker %q is not declared by the DSL", alias.id)
			}
			for _, same := range byName[alias.name] {
				if string(same.ID) != alias.id {
					return nil, fmt.Errorf("generated marker %q conflicts with DSL identity %q for Go name %q", alias.id, same.ID, alias.name)
				}
			}
			if err := r.Register(analyzer.Registration{Ref: analyzer.SymbolRef{PackagePath: authority.Package, PackageName: authority.Package, Name: alias.name}, Kind: analyzeSymbolKind(node.Kind), Identity: analyzer.NewIdentity(node.Namespace.String(), alias.id)}); err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func analyzeSymbolKind(kind semantic.Kind) analyzer.SymbolKind {
	if kind == semantic.Activity {
		return analyzer.KindActivity
	}
	return analyzer.KindEntity
}

func validateAnalyzeAnnotations(result analyzer.Result, authority semantic.IR) error {
	nodes := map[string]semantic.Node{}
	for _, node := range authority.Graph.Nodes() {
		nodes[string(node.ID)] = node
	}
	for _, registration := range result.Registrations {
		node, ok := nodes[registration.Identity.ID]
		if !ok {
			return fmt.Errorf("annotation for %q names identity %q absent from DSL authority", registration.Ref.Name, registration.Identity.ID)
		}
		if analyzeSymbolKind(node.Kind) != registration.Kind || node.Namespace.String() != registration.Identity.Namespace {
			return fmt.Errorf("annotation for %q disagrees with DSL identity %q", registration.Ref.Name, registration.Identity.ID)
		}
	}
	return nil
}

func validateAnalyzeGeneratedSource(model generator.SemanticIR, authority semantic.IR, source []byte) error {
	text := string(source)
	if !strings.Contains(text, "//gooo:generated:") && !strings.Contains(text, "//gooo:slot:") {
		return nil
	}
	if _, err := generator.Generate(model, source); err != nil {
		return err
	}
	regions, slots, err := analyzeMarkers(source)
	if err != nil {
		return err
	}
	nodes := map[string]semantic.Kind{}
	for _, node := range authority.Graph.Nodes() {
		nodes[string(node.ID)] = node.Kind
	}
	seen := map[string]bool{}
	for _, region := range regions {
		if _, ok := nodes[region.id]; !ok {
			return fmt.Errorf("stale generated region identity %q", region.id)
		}
		if seen[region.id] {
			return fmt.Errorf("duplicate generated region identity %q", region.id)
		}
		seen[region.id] = true
	}
	for _, slot := range slots {
		valid := false
		for id, kind := range nodes {
			if kind == semantic.Activity && slot == id+"/implementation" {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("stale protected slot identity %q", slot)
		}
	}
	return nil
}

func analyzeMarkers(source []byte) ([]analyzeGeneratedRegion, []string, error) {
	var regions []analyzeGeneratedRegion
	var slots []string
	lines := strings.Split(string(source), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//gooo:generated:start") {
			id, err := markerID(line, "//gooo:generated:start")
			if err != nil {
				return nil, nil, err
			}
			regions = append(regions, analyzeGeneratedRegion{id: id})
			name := ""
			for _, next := range lines[i+1:] {
				next = strings.TrimSpace(next)
				if next == "" || strings.HasPrefix(next, "//") {
					continue
				}
				fields := strings.Fields(next)
				if len(fields) >= 2 && (fields[0] == "type" || fields[0] == "func") {
					name = strings.Trim(fields[1], "(")
				}
				break
			}
			if name == "" {
				return nil, nil, fmt.Errorf("generated marker %q has no declaration", id)
			}
		}
		if strings.HasPrefix(line, "//gooo:slot:start") {
			id, err := markerID(line, "//gooo:slot:start")
			if err != nil {
				return nil, nil, err
			}
			slots = append(slots, id)
		}
	}
	return regions, slots, nil
}

func analyzeMarkerAliases(source []byte) ([]analyzeMarkerAlias, error) {
	regions, _, err := analyzeMarkers(source)
	if err != nil {
		return nil, err
	}
	aliases := make([]analyzeMarkerAlias, 0, len(regions))
	for _, region := range regions {
		aliases = append(aliases, analyzeMarkerAlias{id: region.id})
	}
	lines := strings.Split(string(source), "\n")
	for i := range aliases {
		for lineIndex, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "//gooo:generated:start") {
				continue
			}
			id, _ := markerID(strings.TrimSpace(line), "//gooo:generated:start")
			if id != aliases[i].id {
				continue
			}
			for _, next := range lines[lineIndex+1:] {
				next = strings.TrimSpace(next)
				if next == "" || strings.HasPrefix(next, "//") {
					continue
				}
				fields := strings.Fields(next)
				if len(fields) >= 2 {
					aliases[i].name = strings.Trim(fields[1], "(")
				}
				break
			}
		}
		if aliases[i].name == "" {
			return nil, fmt.Errorf("generated marker %q has no declaration", aliases[i].id)
		}
	}
	return aliases, nil
}

func marshalAnalyzeDelta(delta analyzer.SemanticNormalizedDelta, authority, observed semantic.IR) ([]byte, error) {
	if delta.SignatureFacts == nil {
		delta.SignatureFacts = []analyzer.NormalizedSignatureFact{}
	}
	if delta.CandidateFacts == nil {
		delta.CandidateFacts = []analyzer.NormalizedCandidateFact{}
	}
	if delta.DeferredFacts == nil {
		delta.DeferredFacts = []analyzer.NormalizedDeferredFact{}
	}
	if delta.DeferredImplementation == nil {
		delta.DeferredImplementation = []analyzer.ImplementationObservation{}
	}
	if delta.DeferredDetails == nil {
		delta.DeferredDetails = []analyzer.DeferredImplementationDetail{}
	}
	if delta.DeferredSlots == nil {
		delta.DeferredSlots = []analyzer.ProtectedSlotObservation{}
	}
	payload, err := json.Marshal(analyzeDeltaOutput{SemanticNormalizedDelta: delta, AuthoritySemanticDigest: authority.StableHash(), ObservedSemanticDigest: observed.StableHash(), SemanticEqual: true, WriteEffect: analyzer.ReconcileNoWrite})
	if err != nil {
		return nil, err
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return nil, errDiagnosticLimit
	}
	return append(payload, '\n'), nil
}
