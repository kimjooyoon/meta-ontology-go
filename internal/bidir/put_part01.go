package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	PutSourceInvalid     = "put.source-invalid"
	PutModelInvalid      = "put.model-invalid"
	PutProvenanceMissing = "put.provenance-missing"
	PutWriteConflict     = "put.write-conflict"
)

// PutError reports a rejected write. NoWrite is true for every PutError: the
// returned document is the original source view, never a partially built one.
type PutError struct {
	Code    string
	NoWrite bool
	Err     error
}

func (e *PutError) Error() string {
	if e == nil {
		return "bidir put failed"
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}
func (e *PutError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Put writes an updated model into a document while preserving source order
// for surviving declarations and explicit relations.
func Put(document Document, updated Model) (Document, error) {
	return putWithTypesAndEntityFieldsSupport(document, updated, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}
func putWithEntityFieldsSupport(document Document, updated Model, support EntityFieldsSupport) (Document, error) {
	return putWithTypesAndEntityFieldsSupport(document, updated, semantic.DefaultTypeRegistry(), support)
}

// PutWithTypes writes a typed model back to the parser-neutral carrier. All
// validation completes before constructing the returned document, preserving
// the no-write guarantee on every rejection.
func PutWithTypes(document Document, updated Model, registry semantic.TypeRegistry) (Document, error) {
	return putWithTypesAndEntityFieldsSupport(document, updated, registry, CurrentEntityFieldsSupport())
}
