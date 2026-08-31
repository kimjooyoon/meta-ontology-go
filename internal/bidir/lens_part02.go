package bidir

import (
	"fmt"
	"strings"
)

func collectDeclarations(model *Model, declarations []Declaration, allowImplicitActivityPorts bool) (map[string]ID, map[ID]struct{}, error) {
	names := make(map[string]ID)
	ids := make(map[ID]struct{}, len(declarations))
	allDeclarations := append([]Declaration(nil), declarations...)
	if allowImplicitActivityPorts {
		allDeclarations = append(allDeclarations, implicitEntityDeclarations(declarations, model.Namespace)...)
	}
	for _, declaration := range allDeclarations {
		if declaration.Kind == "" || strings.TrimSpace(declaration.Name) == "" {
			return nil, nil, fmt.Errorf("declaration %q has empty kind or name", declaration.Name)
		}
		id, err := declarationIdentity(model.Namespace, declaration)
		if err != nil {
			return nil, nil, err
		}
		if _, exists := ids[id]; exists {
			return nil, nil, fmt.Errorf("duplicate declaration ID %q", id)
		}
		if previous, exists := names[referenceKey(model.Namespace, declaration.Name)]; exists {
			return nil, nil, fmt.Errorf("duplicate declaration name %q (%s and %s)", declaration.Name, previous, id)
		}
		ids[id] = struct{}{}
		names[referenceKey(model.Namespace, declaration.Name)] = id
		model.Nodes = append(model.Nodes, Node{ID: id, Kind: declaration.Kind, Name: declaration.Name, Namespace: model.Namespace, Fields: cloneFields(declaration.Fields), Attributes: cloneStringMap(declaration.Attributes), Span: declaration.Span})
	}
	return names, ids, nil
}
func lowerActivityRelations(model *Model, declarations []Declaration, names map[string]ID, ids map[ID]struct{}) error {
	for _, declaration := range declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		activityID, err := declarationIdentity(model.Namespace, declaration)
		if err != nil {
			return err
		}
		for _, reference := range declaration.Inputs {
			entityID, err := resolveReference(reference, model.Namespace, names, ids)
			if err != nil {
				return fmt.Errorf("activity %q input: %w", declaration.Name, err)
			}
			model.Relations = appendUniqueRelation(model.Relations, Relation{Kind: PredicateUsed, Source: activityID, Target: entityID, Span: reference.Span})
		}
		for _, reference := range declaration.Outputs {
			entityID, err := resolveReference(reference, model.Namespace, names, ids)
			if err != nil {
				return fmt.Errorf("activity %q output: %w", declaration.Name, err)
			}
			model.Relations = appendUniqueRelation(model.Relations, Relation{Kind: PredicateWasGeneratedBy, Source: entityID, Target: activityID, Span: reference.Span})
		}
	}
	return nil
}
func lowerExplicitRelations(model *Model, relations []Relation, ids map[ID]struct{}) error {
	for _, relation := range relations {
		relation = relation.normalized()
		if _, exists := ids[relation.Source]; !exists {
			return fmt.Errorf("explicit relation references unknown source %q", relation.Source)
		}
		if _, exists := ids[relation.Target]; !exists {
			return fmt.Errorf("explicit relation references unknown target %q", relation.Target)
		}
		var err error
		model.Relations, err = appendCheckedRelation(model.Relations, relation)
		if err != nil {
			return err
		}
	}
	return nil
}
