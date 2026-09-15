package cycles

import (
	"sort"
	"strings"
)

type edgeInput struct {
	edge      Edge
	subject   ID
	predicate Relation
	object    ID
}

func indexEdges(rawEdges []Edge, nodes map[ID]Node) ([]normalizedEdge, Diagnostics) {
	inputs := make([]edgeInput, 0, len(rawEdges))
	for _, raw := range rawEdges {
		subject, _ := canonicalID(raw.Subject)
		object, _ := canonicalID(raw.Object)
		inputs = append(inputs, edgeInput{edge: raw, subject: subject,
			predicate: relationName(raw), object: object})
	}
	sort.SliceStable(inputs, func(i, j int) bool {
		return edgeInputKey(inputs[i]) < edgeInputKey(inputs[j])
	})

	edges := make([]normalizedEdge, 0, len(inputs))
	diagnostics := make(Diagnostics, 0)
	seen := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		key := edgeIdentityKey(input)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		subject, subjectOK := nodes[input.subject]
		object, objectOK := nodes[input.object]
		if !subjectOK {
			diagnostics = append(diagnostics, unresolvedDiagnostic(input.edge, "subject", input.subject))
		}
		if !objectOK {
			diagnostics = append(diagnostics, unresolvedDiagnostic(input.edge, "object", input.object))
		}
		if !subjectOK || !objectOK {
			continue
		}
		predicate := input.predicate
		_, _, known := expectedKinds(predicate)
		if !known || !validDirection(predicate, subject.Kind, object.Kind) {
			diagnostics = append(diagnostics, Diagnostic{
				Code: IllegalRelationDirection, Subject: input.subject,
				Predicate: predicate, Object: input.object, Span: input.edge.Span,
				Message: directionMessage(input.edge, subject, object, known),
			})
		}
		edges = append(edges, normalizedEdge{subject: input.subject, predicate: predicate,
			object: input.object, span: input.edge.Span, known: known})
	}
	return edges, diagnostics
}
func edgeInputKey(input edgeInput) string {
	return strings.Join([]string{input.subject, string(input.predicate), input.object,
		strings.TrimSpace(input.edge.Subject), strings.TrimSpace(input.edge.Object)}, "\x00")
}
