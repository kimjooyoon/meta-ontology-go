package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
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
	names, ids, err := collectDeclarations(&model, document.Declarations, document.ImplicitActivityPorts)
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
