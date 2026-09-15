package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func declarationFromNode(node Node, model Model, registry semantic.TypeRegistry) (Declaration, error) {
	declaration := Declaration{Kind: node.Kind, ID: node.ID, Name: node.Name, Fields: cloneFields(node.Fields), Attributes: cloneStringMap(node.Attributes), Span: node.Span}
	if declaration.Name == "" {
		declaration.Name = defaultName(node.ID)
	}
	for index, field := range declaration.Fields {
		if err := validateSourceField(field, node.ID, registry); err != nil {
			return Declaration{}, fmt.Errorf("node %q field %d: %w", node.ID, index, err)
		}
	}
	if node.Kind != ActivityKind {
		return declaration, nil
	}
	for _, relation := range model.Relations {
		if relation.Kind == PredicateUsed && relation.Source == node.ID {
			if target, exists := model.node(relation.Target); exists {
				declaration.Inputs = append(declaration.Inputs, Reference{ID: target.ID, Name: target.Name, Namespace: target.Namespace, Span: relation.Span})
			}
		}
		if relation.Kind == PredicateWasGeneratedBy && relation.Target == node.ID {
			if source, exists := model.node(relation.Source); exists {
				declaration.Outputs = append(declaration.Outputs, Reference{ID: source.ID, Name: source.Name, Namespace: source.Namespace, Span: relation.Span})
			}
		}
	}
	sort.SliceStable(declaration.Inputs, func(i, j int) bool {
		return referenceSourceOrderLess(declaration.Inputs[i], declaration.Inputs[j])
	})
	sort.SliceStable(declaration.Outputs, func(i, j int) bool {
		return referenceSourceOrderLess(declaration.Outputs[i], declaration.Outputs[j])
	})
	return declaration, nil
}
func referenceSourceOrderLess(left, right Reference) bool {
	leftValid, rightValid := left.Span.Valid(), right.Span.Valid()
	if leftValid != rightValid {
		return leftValid
	}
	if leftValid {
		if left.Span.File != right.Span.File {
			return left.Span.File < right.Span.File
		}
		if left.Span.Start != right.Span.Start {
			return left.Span.Start < right.Span.Start
		}
		if left.Span.StartLine != right.Span.StartLine {
			return left.Span.StartLine < right.Span.StartLine
		}
		if left.Span.StartColumn != right.Span.StartColumn {
			return left.Span.StartColumn < right.Span.StartColumn
		}
		if left.Span.End != right.Span.End {
			return left.Span.End < right.Span.End
		}
		if left.Span.EndLine != right.Span.EndLine {
			return left.Span.EndLine < right.Span.EndLine
		}
		if left.Span.EndColumn != right.Span.EndColumn {
			return left.Span.EndColumn < right.Span.EndColumn
		}
	}
	return left.ID < right.ID
}
