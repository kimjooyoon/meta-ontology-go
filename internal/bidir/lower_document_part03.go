package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func lowerDocumentNodes(ctx context.Context, ir *semantic.IR, document Document, namespace semantic.Namespace) (map[string]semantic.ID, map[ID]semantic.ID, error) {
	names := make(map[string]semantic.ID)
	ids := make(map[ID]semantic.ID, len(document.Declarations))
	declarations := append(append([]Declaration(nil), document.Declarations...), implicitEntityDeclarations(document.Declarations, namespace.String())...)
	for _, declaration := range declarations {
		if err := checkLowerContext(ctx); err != nil {
			return nil, nil, err
		}
		id, err := declarationIdentity(namespace.String(), declaration)
		if err != nil {
			return nil, nil, err
		}
		semanticID, err := semantic.ParseIdentity(string(id))
		if err != nil {
			return nil, nil, err
		}
		kind, err := semanticKind(declaration.Kind)
		if err != nil {
			return nil, nil, err
		}
		node, err := semantic.NewNode(kind, semanticID, namespace, declaration.Name)
		if err != nil {
			return nil, nil, err
		}
		if err := bindSemanticValueProgram(declaration, &node); err != nil {
			return nil, nil, err
		}
		if len(declaration.Fields) > 0 {
			if kind != semantic.Entity {
				return nil, nil, fmt.Errorf("declaration %q: fields are only valid on Entity nodes", declaration.Name)
			}
			node.Fields = make([]semantic.Field, len(declaration.Fields))
			for index, field := range declaration.Fields {
				semanticField, err := field.semantic()
				if err != nil {
					return nil, nil, fmt.Errorf("declaration %q field %d: %w", declaration.Name, index, err)
				}
				if semanticField.Parent != semanticID {
					return nil, nil, fmt.Errorf("declaration %q field %d: %w: field %s parent is %s, want %s", declaration.Name, index, ErrInvalidField, semanticField.ID, semanticField.Parent, semanticID)
				}
				node.Fields[index] = semanticField
			}
		}
		node.Span = toSemanticSpan(declaration.Span)
		if err := ir.AddNode(node); err != nil {
			return nil, nil, err
		}
		if _, exists := ids[id]; exists {
			return nil, nil, fmt.Errorf("duplicate declaration ID %q", id)
		}
		ids[id] = semanticID
		names[referenceKey(namespace.String(), declaration.Name)] = semanticID
	}
	return names, ids, nil
}
func semanticIRHasFields(ir semantic.IR) bool {
	for _, node := range ir.Graph.Nodes() {
		if len(node.Fields) > 0 {
			return true
		}
	}
	return false
}
