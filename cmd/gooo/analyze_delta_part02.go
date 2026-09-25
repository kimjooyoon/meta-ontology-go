package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
