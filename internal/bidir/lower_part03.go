package bidir

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
)

func syntaxDeclarations(file *syntax.File) []syntax.Declaration {
	if len(file.Declarations) > 0 {
		return file.Declarations
	}
	return file.Decls
}
func adaptSyntaxDeclaration(ctx context.Context, declaration syntax.Declaration) (Declaration, error) {
	if err := checkLowerContext(ctx); err != nil {
		return Declaration{}, err
	}
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		if len(value.Fields) > 0 && strings.TrimSpace(value.ID) == "" {
			return Declaration{}, fmt.Errorf("entity %q field parent ID is required", value.Name)
		}
		fields := make([]Field, 0, len(value.Fields))
		for index, field := range value.Fields {
			if err := checkLowerContext(ctx); err != nil {
				return Declaration{}, err
			}
			adapted, err := adaptSyntaxField(value.ID, field)
			if err != nil {
				return Declaration{}, fmt.Errorf("entity %q field %d: %w", value.Name, index, err)
			}
			fields = append(fields, adapted)
		}
		return Declaration{Kind: EntityKind, ID: ID(value.ID), Name: value.Name, Fields: fields, Span: toSourceSpan(value.Span)}, nil
	case *syntax.ActivityDecl:
		return adaptSyntaxActivity(ctx, value)
	default:
		return Declaration{}, fmt.Errorf("unsupported syntax declaration %T", declaration)
	}
}
func adaptSyntaxField(parent string, field syntax.FieldDecl) (Field, error) {
	typeRef, typeRefUse, err := parseSyntaxTypeRef(field.TypeRef.Spelling)
	if err != nil {
		return Field{}, err
	}
	typeRefUse.Span = toSourceSpan(field.TypeRef.Span)
	return Field{
		ID:              ID(field.ID),
		Parent:          ID(parent),
		Name:            field.Name,
		TypeRef:         typeRef,
		TypeRefUse:      typeRefUse,
		Origin:          FieldOriginSource,
		Presence:        FieldPresence(field.Presence),
		Cardinality:     FieldCardinality(field.Cardinality),
		Span:            toSourceSpan(field.Span),
		IDSpan:          toSourceSpan(field.IDSpan),
		NameSpan:        toSourceSpan(field.NameSpan),
		TypeRefSpan:     toSourceSpan(field.TypeRef.Span),
		PresenceSpan:    toSourceSpan(field.PresenceSpan),
		CardinalitySpan: toSourceSpan(field.CardinalitySpan),
	}, nil
}
