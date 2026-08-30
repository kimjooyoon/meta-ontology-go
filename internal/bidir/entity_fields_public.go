package bidir

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// EntityFieldsV1Support is the explicit public activation boundary for the
// profile. Ordinary Get, Put, and Lower keep their deferred defaults.
func EntityFieldsV1Support() EntityFieldsSupport { return syntax.EntityFieldsV1Support() }

// DocumentFromSyntaxWithEntityFieldsSupport adapts a profile-bound AST.
func DocumentFromSyntaxWithEntityFieldsSupport(file *syntax.File, support EntityFieldsSupport) (Document, error) {
	return documentFromSyntaxWithEntityFieldsSupport(file, support)
}

// LowerContextWithEntityFieldsSupport lowers a profile-bound AST.
func LowerContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithEntityFieldsSupport(ctx, file, support)
}

func LowerDocumentWithEntityFieldsSupport(document Document, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerDocumentWithEntityFieldsSupport(document, support)
}

// GetWithEntityFieldsSupport reads a profile-bound document.
func GetWithEntityFieldsSupport(document Document, support EntityFieldsSupport) (Model, error) {
	return getWithEntityFieldsSupport(document, support)
}

// PutWithEntityFieldsSupport writes a profile-bound model without partial data.
func PutWithEntityFieldsSupport(document Document, updated Model, support EntityFieldsSupport) (Document, error) {
	return putWithEntityFieldsSupport(document, updated, support)
}

// CheckGetPutWithEntityFieldsSupport verifies the profile-bound BX roundtrip.
func CheckGetPutWithEntityFieldsSupport(document Document, support EntityFieldsSupport) error {
	model, err := GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return err
	}
	written, err := PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		return err
	}
	observed, err := GetWithEntityFieldsSupport(written, support)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return ErrInvalidField
	}
	return nil
}

func CheckPutGetWithEntityFieldsSupport(document Document, model Model, support EntityFieldsSupport) error {
	written, err := PutWithEntityFieldsSupport(document, model, support)
	if err != nil {
		return err
	}
	observed, err := GetWithEntityFieldsSupport(written, support)
	if err != nil {
		return err
	}
	if !SemanticEquivalent(model, observed) {
		return ErrInvalidField
	}
	return nil
}
