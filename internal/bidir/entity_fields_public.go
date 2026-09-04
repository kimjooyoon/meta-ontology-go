package bidir

import (
	"context"
	"fmt"
	"reflect"

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

// DocumentFromSyntaxWithImplicitActivityPorts opts a source view into the
// legacy surface where activity ports are declarations by reference.
func DocumentFromSyntaxWithImplicitActivityPorts(file *syntax.File, support EntityFieldsSupport) (Document, error) {
	document, err := documentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return Document{}, err
	}
	document.ImplicitActivityPorts = true
	return document, nil
}

// LowerContextWithEntityFieldsSupport lowers a profile-bound AST.
func LowerContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithEntityFieldsSupport(ctx, file, support)
}

// LowerContextWithImplicitActivityPorts opts a source AST into the legacy
// surface where activity ports are declarations by reference.
func LowerContextWithImplicitActivityPorts(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupportAndImplicitActivityPorts(ctx, file, semantic.DefaultTypeRegistry(), support, true)
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
	if !DocumentEquivalent(document, written) {
		return fmt.Errorf("Get-Put violated: source document changed after unchanged write-back")
	}
	if !SemanticEquivalent(model, observed) {
		return fmt.Errorf("Get-Put violated: semantic model changed after unchanged write-back")
	}
	return nil
}

// DocumentEquivalent is the strict source-preserving comparison for the
// profile-bound Get-Put law. It includes declaration order, stable IDs, and
// every source span instead of comparing only normalized semantic nodes.
// Derived activity IDs and resolved type IDs are canonicalized because Put
// fills those values when reconstructing the parser-neutral carrier.
func DocumentEquivalent(left, right Document) bool {
	return reflect.DeepEqual(documentEvidence(left), documentEvidence(right))
}

type canonicalDocumentEvidence struct {
	Package      string
	Namespace    string
	Declarations []declarationEvidence
	Policies     []semantic.Policy
}

type declarationEvidence struct {
	Kind       Kind
	ID         ID
	Name       string
	Fields     []fieldEvidence
	Inputs     []referenceEvidence
	Outputs    []referenceEvidence
	Attributes map[string]string
	Span       SourceSpan
}

type fieldEvidence struct {
	ID, Parent                    ID
	Name                          string
	TypeRef                       TypeRef
	TypeRefPresentation           TypeRefUse
	Origin                        FieldOrigin
	Presence                      FieldPresence
	Cardinality                   FieldCardinality
	Span, IDSpan                  SourceSpan
	NameSpan, TypeRefSpan         SourceSpan
	PresenceSpan, CardinalitySpan SourceSpan
}

type referenceEvidence struct {
	ID        ID
	Namespace string
	Name      string
	Span      SourceSpan
}

func documentEvidence(document Document) canonicalDocumentEvidence {
	result := canonicalDocumentEvidence{Package: document.Package, Namespace: document.Namespace,
		Policies: append([]semantic.Policy(nil), document.Policies...)}
	idsByName := make(map[string]ID, len(document.Declarations))
	for _, declaration := range document.Declarations {
		id, _ := declarationIdentity(document.Namespace, declaration)
		idsByName[declaration.Name] = id
	}
	for _, declaration := range document.Declarations {
		id, _ := declarationIdentity(document.Namespace, declaration)
		item := declarationEvidence{Kind: declaration.Kind, ID: id, Name: declaration.Name, Attributes: cloneStringMap(declaration.Attributes), Span: declaration.Span}
		for _, field := range declaration.Fields {
			item.Fields = append(item.Fields, fieldEvidence{ID: field.ID, Parent: field.Parent, Name: field.Name, TypeRef: field.TypeRef, TypeRefPresentation: canonicalTypeRefPresentation(field.TypeRefUse), Origin: field.Origin, Presence: field.Presence, Cardinality: field.Cardinality, Span: field.Span, IDSpan: field.IDSpan, NameSpan: field.NameSpan, TypeRefSpan: field.TypeRefSpan, PresenceSpan: field.PresenceSpan, CardinalitySpan: field.CardinalitySpan})
		}
		for _, reference := range declaration.Inputs {
			item.Inputs = append(item.Inputs, canonicalReferenceEvidence(reference, idsByName, document.Namespace))
		}
		for _, reference := range declaration.Outputs {
			item.Outputs = append(item.Outputs, canonicalReferenceEvidence(reference, idsByName, document.Namespace))
		}
		result.Declarations = append(result.Declarations, item)
	}
	return result
}

func canonicalReferenceEvidence(reference Reference, idsByName map[string]ID, namespace string) referenceEvidence {
	id := reference.ID
	if id == "" {
		id = idsByName[reference.Name]
	}
	refNamespace := reference.Namespace
	if refNamespace == "" {
		refNamespace = namespace
	}
	if id == "" && refNamespace == namespace {
		id, _ = declarationIdentity(refNamespace, Declaration{Kind: EntityKind, Name: reference.Name})
	}
	return referenceEvidence{ID: id, Name: reference.Name, Namespace: refNamespace, Span: reference.Span}
}

func canonicalTypeRefPresentation(use TypeRefUse) TypeRefUse {
	use.ResolvedID = ""
	return use
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
