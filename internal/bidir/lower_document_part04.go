package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
func lowerDocumentContracts(ctx context.Context, ir *semantic.IR, document Document, namespace semantic.Namespace, ids map[ID]semantic.ID, names map[string]semantic.ID) error {
	for _, declaration := range document.Declarations {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
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
		if err := lowerTypedContractReferences(ctx, ir, activityID, declaration, namespace, ids, names); err != nil {
			return err
		}
	}
	return nil
}
func lowerTypedContractReferences(ctx context.Context, ir *semantic.IR, activityID semantic.ID, declaration Declaration, namespace semantic.Namespace, ids map[ID]semantic.ID, names map[string]semantic.ID) error {
	for _, reference := range declaration.Inputs {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
		id, err := resolveSemanticReference(reference, namespace, ids, names)
		if err != nil {
			return fmt.Errorf("activity %q input: %w", declaration.Name, err)
		}
		fact := semantic.NewUsedFact(activityID, id).WithSpan(toSemanticSpan(reference.Span))
		if err := ir.AddFact(fact); err != nil {
			return fmt.Errorf("activity %q input: %w", declaration.Name, err)
		}
	}
	for _, reference := range declaration.Outputs {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
		id, err := resolveSemanticReference(reference, namespace, ids, names)
		if err != nil {
			return fmt.Errorf("activity %q output: %w", declaration.Name, err)
		}
		fact := semantic.NewWasGeneratedByFact(id, activityID).WithSpan(toSemanticSpan(reference.Span))
		if err := ir.AddFact(fact); err != nil {
			return fmt.Errorf("activity %q output: %w", declaration.Name, err)
		}
	}
	return nil
}
