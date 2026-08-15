package bidir

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// DSLAdapter defines the parser-neutral source boundary.
type DSLAdapter interface {
	Decode(source []byte) (Document, error)
	Encode(document Document) ([]byte, error)
}

// GoFactAdapter defines the source-backed Go analysis boundary.
type GoFactAdapter interface {
	Analyze(source []byte) (FactDelta, error)
}

// Get lowers a parser-neutral document into the canonical generic model.
func Get(document Document) (Model, error) {
	return getWithTypesAndEntityFieldsSupport(document, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}

// getWithEntityFieldsSupport is the explicit support-boundary variant used by
// tests that inject the exact profile-bound SUPPORTED state.
func getWithEntityFieldsSupport(document Document, support EntityFieldsSupport) (Model, error) {
	return getWithTypesAndEntityFieldsSupport(document, semantic.DefaultTypeRegistry(), support)
}

// GetWithTypes lowers a parser-neutral document into the generic model while
// resolving every latent field TypeRef through registry.
func GetWithTypes(document Document, registry semantic.TypeRegistry) (Model, error) {
	return getWithTypesAndEntityFieldsSupport(document, registry, CurrentEntityFieldsSupport())
}

// getWithTypesAndEntityFieldsSupport lowers with an explicitly injected
// EntityFields support binding and never returns a partial model.
func getWithTypesAndEntityFieldsSupport(document Document, registry semantic.TypeRegistry, support EntityFieldsSupport) (Model, error) {
	if err := entityFieldsActivation(support, documentHasFields(document), firstDocumentFieldSpan(document)); err != nil {
		return Model{}, err
	}
	if err := validateDocumentSpans(document); err != nil {
		return Model{}, err
	}
	model := Model{Package: document.Package, Namespace: document.Namespace}
	if strings.TrimSpace(model.Namespace) == "" {
		model.Namespace = "gooo"
	}
	names, ids, err := collectDeclarations(&model, document.Declarations)
	if err != nil {
		return Model{}, err
	}
	if err := lowerActivityRelations(&model, document.Declarations, names, ids); err != nil {
		return Model{}, err
	}
	if err := lowerExplicitRelations(&model, document.Relations, ids); err != nil {
		return Model{}, err
	}
	model.Normalize()
	if err := normalizeModelFields(&model, registry); err != nil {
		return Model{}, classifyEntityFieldsModelError(err, firstModelFieldSpan(model.Nodes))
	}
	if err := validateEntityFieldsModel(model.Nodes, registry, support); err != nil {
		return Model{}, err
	}
	if err := model.ValidateWithTypes(registry); err != nil {
		return Model{}, err
	}
	return model, nil
}

func collectDeclarations(model *Model, declarations []Declaration) (map[string]ID, map[ID]struct{}, error) {
	names := make(map[string]ID)
	ids := make(map[ID]struct{}, len(declarations))
	for _, declaration := range declarations {
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
