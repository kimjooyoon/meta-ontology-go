package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
func lowerDocumentRelations(ctx context.Context, ir *semantic.IR, relations []Relation) error {
	for _, relation := range relations {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
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

func lowerDocumentRuntimeBindings(ctx context.Context, ir *semantic.IR, document Document, namespace semantic.Namespace, ids map[ID]semantic.ID, names map[string]semantic.ID) error {
	for index, binding := range document.RuntimeBindings {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
		producer, err := resolveSemanticReference(binding.Producer.Activity, namespace, ids, names)
		if err != nil {
			return fmt.Errorf("runtime binding %d producer: %w", index, err)
		}
		consumer, err := resolveSemanticReference(binding.Consumer.Activity, namespace, ids, names)
		if err != nil {
			return fmt.Errorf("runtime binding %d consumer: %w", index, err)
		}
		ir.RuntimeBindings = append(ir.RuntimeBindings, semantic.RuntimeBinding{
			Schema:           semantic.RuntimeBindingSchema,
			ProducerActivity: producer,
			ProducerPort:     binding.Producer.Port.Name,
			ConsumerActivity: consumer,
			ConsumerPort:     binding.Consumer.Port.Name,
			Entity:           semantic.ID(binding.Entity),
			Span:             toSemanticSpan(binding.Span),
		})
	}
	return nil
}

func validateLoweredContext(ctx context.Context, ir semantic.IR) error {
	for range ir.Graph.Nodes() {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
	}
	for range ir.Graph.AllFacts() {
		if err := checkLowerContext(ctx); err != nil {
			return err
		}
	}
	return checkLowerContext(ctx)
}
