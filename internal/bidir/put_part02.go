package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func putWithTypesAndEntityFieldsSupport(document Document, updated Model, registry semantic.TypeRegistry, support EntityFieldsSupport) (Document, error) {
	hasFields := documentHasFields(document) || modelHasFields(updated.Nodes)
	if err := entityFieldsActivation(support, hasFields, firstPutFieldSpan(document, updated)); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	source, err := getWithTypesAndEntityFieldsSupport(document, registry, support)
	if err != nil {
		return document, putError(PutSourceInvalid, err)
	}
	updated = updated.Normalized()
	if err := validateFieldParentStability(source, updated); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	if err := validateFieldOrderStability(source, updated); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	if err := validateFieldSourceSnapshot(source, updated); err != nil {
		return document, putError(PutProvenanceMissing, err)
	}
	if err := validatePutProvenance(source, updated); err != nil {
		return document, putError(PutProvenanceMissing, err)
	}
	if err := validateEntityFieldsModel(updated.Nodes, registry, support); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	if err := updated.ValidateWithTypes(registry); err != nil {
		return document, putError(PutModelInvalid, err)
	}
	result := Document{Package: document.Package, Namespace: document.Namespace,
		Policies: append([]semantic.Policy(nil), document.Policies...), ImplicitActivityPorts: document.ImplicitActivityPorts}
	if result.Package == "" {
		result.Package = updated.Package
	}
	if result.Namespace == "" {
		result.Namespace = updated.Namespace
	}
	if result.Namespace == "" {
		result.Namespace = "gooo"
	}
	nodes := nodeMap(updated.Nodes)
	declarationIDs, err := appendSurvivingDeclarations(&result, document.Declarations, nodes, updated, registry)
	if err != nil {
		return document, putError(PutWriteConflict, err)
	}
	if err := appendNewDeclarations(&result, updated, nodes, declarationIDs, registry); err != nil {
		return document, putError(PutWriteConflict, err)
	}
	appendUpdatedRelations(&result, document.Relations, updated)
	return result, nil
}
func firstPutFieldSpan(document Document, updated Model) SourceSpan {
	if span := firstDocumentFieldSpan(document); span.Valid() {
		return span
	}
	return firstModelFieldSpan(updated.Nodes)
}
