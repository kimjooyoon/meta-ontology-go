package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func firstDocumentFieldSpan(document Document) SourceSpan {
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) != 0 {
			return declaration.Fields[0].Span
		}
	}
	return SourceSpan{}
}
func validateEntityFieldsDocument(document Document, namespace string, registry semantic.TypeRegistry, support EntityFieldsSupport) error {
	if err := entityFieldsActivation(support, documentHasFields(document), firstDocumentFieldSpan(document)); err != nil {
		return err
	}
	if !documentHasFields(document) {
		return nil
	}
	for _, declaration := range document.Declarations {
		if len(declaration.Fields) == 0 {
			continue
		}
		if declaration.Kind != EntityKind {
			return entityFieldsError(EntityFieldsWrongParentDiagnostic, "fields are only valid on Entity declarations", declaration.Span, ErrInvalidField)
		}
		owner, err := declarationIdentity(namespace, declaration)
		if err != nil {
			return err
		}
		if err := validateEntityFieldSequence(owner, declaration.Fields); err != nil {
			return err
		}
		for _, field := range declaration.Fields {
			if err := validateSourceField(field, owner, registry); err != nil {
				return err
			}
			if err := validateEntityFieldsProfileField(field, registry); err != nil {
				return err
			}
		}
	}
	return nil
}
func validateEntityFieldSequence(owner ID, fields []Field) error {
	if len(fields) < 2 {
		return nil
	}
	file := fields[0].Span.File
	for index := 1; index < len(fields); index++ {
		if fields[index].Span.File != file || fields[index-1].Span.Start > fields[index].Span.Start {
			return entityFieldsError(EntityFieldsIllegalReorderDiagnostic, fmt.Sprintf("field order for entity %s is not source ordered", owner), fields[index].Span, ErrInvalidField)
		}
	}
	return nil
}
