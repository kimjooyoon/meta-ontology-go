package analyzer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func addRegisteredNodes(ir *semantic.IR, registrations []Registration) error {
	ordered := append([]Registration(nil), registrations...)
	sortRegistrations(ordered)
	for _, registration := range ordered {
		id, err := semantic.ParseIdentity(registration.Identity.ID)
		if err != nil {
			return adapterError(AdapterUnknownEndpoint, "", registration.Identity.ID, err.Error())
		}
		namespace, err := semantic.ParseNamespace(registration.Identity.Namespace)
		if err != nil {
			return adapterError(AdapterUnknownEndpoint, "", registration.Identity.ID, err.Error())
		}
		kind, err := registrationKind(registration.Kind)
		if err != nil {
			return err
		}
		name := registration.Ref.Name
		node, err := semantic.NewNode(kind, id, namespace, name)
		if err != nil {
			return adapterError(AdapterUnknownEndpoint, "", registration.Identity.ID, err.Error())
		}
		node.Span = semanticSpan(registration.Span)
		if existing, ok := ir.Graph.Node(id); ok {
			if existing.Kind != kind || existing.Namespace != namespace {
				return adapterError(AdapterEndpointKind, "", id.String(), "registered node conflicts with semantic graph")
			}
			continue
		}
		if err := ir.AddNode(node); err != nil {
			return err
		}
	}
	return nil
}

func registrationKind(kind SymbolKind) (semantic.Kind, error) {
	switch kind {
	case KindActivity:
		return semantic.Activity, nil
	case KindEntity:
		return semantic.Entity, nil
	default:
		return "", adapterError(AdapterEndpointKind, "", "", fmt.Sprintf("unsupported registration kind %q", kind))
	}
}

func mapFact(graph semantic.Graph, subject, object Identity, mapping RelationMapping, span Span) (semantic.Fact, error) {
	subjectID, subjectNode, err := resolveEndpoint(graph, subject)
	if err != nil {
		return semantic.Fact{}, err
	}
	objectID, objectNode, err := resolveEndpoint(graph, object)
	if err != nil {
		return semantic.Fact{}, err
	}
	if subjectNode.Kind != mapping.SourceSubjectKind {
		return semantic.Fact{}, adapterError(AdapterEndpointKind, mapping.Source, subjectID.String(), fmt.Sprintf("want source subject %s, got %s", mapping.SourceSubjectKind, subjectNode.Kind))
	}
	if objectNode.Kind != mapping.SourceObjectKind {
		return semantic.Fact{}, adapterError(AdapterEndpointKind, mapping.Source, objectID.String(), fmt.Sprintf("want source object %s, got %s", mapping.SourceObjectKind, objectNode.Kind))
	}
	if mapping.Reverse {
		subjectID, objectID = objectID, subjectID
	}
	if err := mapping.Predicate.ValidateKinds(nodeKind(graph, subjectID), nodeKind(graph, objectID)); err != nil {
		return semantic.Fact{}, adapterError(AdapterEndpointKind, mapping.Source, subjectID.String(), err.Error())
	}
	return semantic.NewFact(subjectID, mapping.Predicate, objectID).WithSpan(semanticSpan(span)), nil
}

func resolveEndpoint(graph semantic.Graph, identity Identity) (semantic.ID, semantic.Node, error) {
	id, err := semantic.ParseIdentity(identity.ID)
	if err != nil {
		return "", semantic.Node{}, adapterError(AdapterUnknownEndpoint, "", identity.ID, err.Error())
	}
	node, ok := graph.Node(id)
	if !ok {
		return "", semantic.Node{}, adapterError(AdapterUnknownEndpoint, "", id.String(), "endpoint is not declared in the semantic graph")
	}
	return id, node, nil
}

func nodeKind(graph semantic.Graph, id semantic.ID) semantic.Kind {
	node, _ := graph.Node(id)
	return node.Kind
}

func semanticSpan(span Span) semantic.Span {
	return semantic.Span{
		File:  span.Filename,
		Start: semantic.Position{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   semantic.Position{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}
