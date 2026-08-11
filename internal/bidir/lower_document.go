package bidir

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// LowerDocument lowers the parser-neutral view into the current semantic IR.
func LowerDocument(document Document) (semantic.IR, error) {
	namespaceText := strings.TrimSpace(document.Namespace)
	if namespaceText == "" {
		namespaceText = "gooo"
	}
	namespace, err := semantic.ParseNamespace(namespaceText)
	if err != nil {
		return semantic.IR{}, err
	}
	ir := semantic.NewIR(document.Package, namespace)
	names, ids, err := lowerDocumentNodes(&ir, document, namespace)
	if err != nil {
		return semantic.IR{}, err
	}
	if err := lowerDocumentContracts(&ir, document, namespace, ids, names); err != nil {
		return semantic.IR{}, err
	}
	if err := lowerDocumentRelations(&ir, document.Relations); err != nil {
		return semantic.IR{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, err
	}
	return ir, nil
}

func lowerDocumentNodes(ir *semantic.IR, document Document, namespace semantic.Namespace) (map[string]semantic.ID, map[ID]semantic.ID, error) {
	names := make(map[string]semantic.ID)
	ids := make(map[ID]semantic.ID, len(document.Declarations))
	for _, declaration := range document.Declarations {
		if len(declaration.Attributes) > 0 {
			return nil, nil, fmt.Errorf("semantic IR does not support declaration attributes")
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

func semanticKind(kind Kind) (semantic.Kind, error) {
	switch kind {
	case EntityKind:
		return semantic.Entity, nil
	case ActivityKind:
		return semantic.Activity, nil
	case AgentKind:
		return semantic.Agent, nil
	default:
		return "", fmt.Errorf("unsupported semantic kind %q", kind)
	}
}

func lowerDocumentContracts(ir *semantic.IR, document Document, namespace semantic.Namespace, ids map[ID]semantic.ID, names map[string]semantic.ID) error {
	for _, declaration := range document.Declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		activityID := ids[ID(declaration.ID)]
		if activityID == "" {
			generated, err := declarationIdentity(namespace.String(), declaration)
			if err != nil {
				return err
			}
			activityID = ids[generated]
		}
		contract := semantic.ActivityContract{Activity: activityID, Span: toSemanticSpan(declaration.Span)}
		for _, reference := range declaration.Inputs {
			id, err := resolveSemanticReference(reference, namespace, ids, names)
			if err != nil {
				return fmt.Errorf("activity %q input: %w", declaration.Name, err)
			}
			contract.Inputs = append(contract.Inputs, id)
		}
		for _, reference := range declaration.Outputs {
			id, err := resolveSemanticReference(reference, namespace, ids, names)
			if err != nil {
				return fmt.Errorf("activity %q output: %w", declaration.Name, err)
			}
			contract.Outputs = append(contract.Outputs, id)
		}
		if err := ir.AddActivityContract(contract); err != nil {
			return err
		}
	}
	return nil
}

func resolveSemanticReference(reference Reference, namespace semantic.Namespace, ids map[ID]semantic.ID, names map[string]semantic.ID) (semantic.ID, error) {
	if reference.ID != "" {
		id, err := semantic.ParseIdentity(string(reference.ID))
		if err != nil {
			return "", err
		}
		if _, exists := ids[reference.ID]; !exists {
			return "", fmt.Errorf("unknown semantic ID %q", reference.ID)
		}
		return id, nil
	}
	refNamespace := reference.Namespace
	if refNamespace == "" {
		refNamespace = namespace.String()
	}
	id, exists := names[referenceKey(refNamespace, reference.Name)]
	if !exists {
		return "", fmt.Errorf("unknown declaration %q in namespace %q", reference.Name, refNamespace)
	}
	return id, nil
}

func lowerDocumentRelations(ir *semantic.IR, relations []Relation) error {
	for _, relation := range relations {
		predicate, ok := semanticPredicate(relation.Kind)
		if !ok {
			return fmt.Errorf("predicate %q is not representable in semantic IR", relation.Kind)
		}
		if len(relation.Attributes) > 0 {
			return fmt.Errorf("semantic IR does not support relation attributes")
		}
		subject, err := semantic.ParseIdentity(string(relation.Source))
		if err != nil {
			return err
		}
		object, err := semantic.ParseIdentity(string(relation.Target))
		if err != nil {
			return err
		}
		if err := ir.AddFact(semantic.NewFact(subject, predicate, object).WithSpan(toSemanticSpan(relation.Span))); err != nil {
			return err
		}
	}
	return nil
}

func semanticPredicate(predicate Predicate) (semantic.Relation, bool) {
	switch predicate {
	case PredicateUsed:
		return semantic.Used, true
	case PredicateWasGeneratedBy:
		return semantic.WasGeneratedBy, true
	case PredicateWasDerivedFrom:
		return semantic.WasDerivedFrom, true
	default:
		return "", false
	}
}
