package bidir

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// ErrLowerCanceled is returned for cancellation and deadline expiration.
// It is intentionally stable so callers do not need to inspect context text.
var ErrLowerCanceled = errors.New("bidir lowering canceled")

// Lower parses the current syntax AST boundary into semantic IR.
func Lower(file *syntax.File) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}

func lowerWithEntityFieldsSupport(file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, semantic.DefaultTypeRegistry(), support)
}

// LowerWithTypes lowers the syntax carrier and resolves latent field TypeRefs
// through the supplied semantic registry.
func LowerWithTypes(file *syntax.File, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerWithTypesAndEntityFieldsSupport(file, registry, CurrentEntityFieldsSupport())
}

func lowerWithTypesAndEntityFieldsSupport(file *syntax.File, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(context.Background(), file, registry, support)
}

// LowerContext lowers without mutating file and never returns a partial IR.
func LowerContext(ctx context.Context, file *syntax.File) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, semantic.DefaultTypeRegistry(), CurrentEntityFieldsSupport())
}

// LowerContextWithTypes is the cancellable typed syntax lowerer.
func LowerContextWithTypes(ctx context.Context, file *syntax.File, registry semantic.TypeRegistry) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, registry, CurrentEntityFieldsSupport())
}

func lowerContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (semantic.IR, error) {
	return lowerContextWithTypesAndEntityFieldsSupport(ctx, file, semantic.DefaultTypeRegistry(), support)
}

func lowerContextWithTypesAndEntityFieldsSupport(ctx context.Context, file *syntax.File, registry semantic.TypeRegistry, support EntityFieldsSupport) (semantic.IR, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return semantic.IR{}, err
	}
	document, err := documentFromSyntaxContextWithEntityFieldsSupport(ctx, file, support)
	if err != nil {
		return semantic.IR{}, err
	}
	return lowerDocumentContextWithTypesAndEntityFieldsSupport(ctx, document, registry, support)
}

func nonNilLowerContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func checkLowerContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrLowerCanceled
	default:
		return nil
	}
}

// DocumentFromSyntax adapts syntax without making the generic lens depend on
// parser implementation details.
func DocumentFromSyntax(file *syntax.File) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(context.Background(), file, CurrentEntityFieldsSupport())
}

func documentFromSyntaxWithEntityFieldsSupport(file *syntax.File, support EntityFieldsSupport) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(context.Background(), file, support)
}

// DocumentFromSyntaxContext is the cancellable syntax adapter.
func DocumentFromSyntaxContext(ctx context.Context, file *syntax.File) (Document, error) {
	return documentFromSyntaxContextWithEntityFieldsSupport(ctx, file, CurrentEntityFieldsSupport())
}

func documentFromSyntaxContextWithEntityFieldsSupport(ctx context.Context, file *syntax.File, support EntityFieldsSupport) (Document, error) {
	ctx = nonNilLowerContext(ctx)
	if err := checkLowerContext(ctx); err != nil {
		return Document{}, err
	}
	if file == nil || file.Package == nil {
		return Document{}, fmt.Errorf("package is required")
	}
	if file.Namespace == nil || file.Namespace.Name == "" {
		return Document{}, fmt.Errorf("namespace is required")
	}
	if err := entityFieldsActivation(support, syntaxFileHasFields(file), firstSyntaxFieldSpan(file)); err != nil {
		return Document{}, err
	}
	document := Document{Package: file.Package.Name, Namespace: file.Namespace.Name}
	for _, declaration := range syntaxDeclarations(file) {
		if err := checkLowerContext(ctx); err != nil {
			return Document{}, err
		}
		adapted, err := adaptSyntaxDeclaration(ctx, declaration)
		if err != nil {
			return Document{}, err
		}
		document.Declarations = append(document.Declarations, adapted)
	}
	if err := validateEntityFieldsDocument(document, document.Namespace, semantic.DefaultTypeRegistry(), support); err != nil {
		return Document{}, err
	}
	return document, nil
}

func syntaxFileHasFields(file *syntax.File) bool {
	if file == nil {
		return false
	}
	for _, declaration := range syntaxDeclarations(file) {
		entity, ok := declaration.(*syntax.EntityDecl)
		if ok && len(entity.Fields) > 0 {
			return true
		}
	}
	return false
}

func firstSyntaxFieldSpan(file *syntax.File) SourceSpan {
	if file == nil {
		return SourceSpan{}
	}
	for _, declaration := range syntaxDeclarations(file) {
		entity, ok := declaration.(*syntax.EntityDecl)
		if ok && len(entity.Fields) > 0 {
			return toSourceSpan(entity.Fields[0].Span)
		}
	}
	return SourceSpan{}
}

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

func parseSyntaxTypeRef(spelling string) (semantic.TypeRef, TypeRefUse, error) {
	raw := strings.TrimSpace(spelling)
	if raw == "" {
		return semantic.TypeRef{}, TypeRefUse{}, fmt.Errorf("%w: type reference spelling is empty", ErrInvalidField)
	}
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "urn:") {
		id, err := semantic.ParseIdentity(raw)
		if err != nil {
			return semantic.TypeRef{}, TypeRefUse{}, fmt.Errorf("%w: type reference identity: %v", ErrInvalidField, err)
		}
		return semantic.TypeRef{ID: id}, TypeRefUse{Form: TypeRefFormStableID, Spelling: id.String(), ResolvedID: ID(id)}, nil
	}
	ref, err := parseLookupTypeRef(raw)
	if err != nil {
		return semantic.TypeRef{}, TypeRefUse{}, err
	}
	return ref, TypeRefUse{Form: TypeRefFormLookup, Spelling: lookupTypeRefSpelling(ref)}, nil
}

func parseLookupTypeRef(raw string) (semantic.TypeRef, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference spelling is empty", ErrInvalidField)
	}
	parts := strings.Split(raw, ":")
	ref := semantic.TypeRef{}
	switch len(parts) {
	case 1:
		ref.Name = strings.TrimSpace(parts[0])
	case 2:
		namespace, err := semantic.ParseNamespace(strings.TrimSpace(parts[0]))
		if err != nil {
			return semantic.TypeRef{}, fmt.Errorf("%w: type reference namespace: %v", ErrInvalidField, err)
		}
		ref.Namespace = namespace
		ref.Name = strings.TrimSpace(parts[1])
	default:
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference spelling %q is not representable", ErrInvalidField, raw)
	}
	if err := ref.Validate(); err != nil {
		return semantic.TypeRef{}, fmt.Errorf("%w: type reference: %v", ErrInvalidField, err)
	}
	return ref, nil
}

func adaptSyntaxActivity(ctx context.Context, activity *syntax.ActivityDecl) (Declaration, error) {
	if err := checkLowerContext(ctx); err != nil {
		return Declaration{}, err
	}
	declaration := Declaration{Kind: ActivityKind, Name: activity.Name, Span: toSourceSpan(activity.Span)}
	if len(activity.Inputs) == 0 && len(activity.Parameters) != 0 {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Parameters; canonical Inputs is required", activity.Name)
	}
	for _, input := range activity.Inputs {
		if err := checkLowerContext(ctx); err != nil {
			return Declaration{}, err
		}
		declaration.Inputs = append(declaration.Inputs, Reference{Name: input.Name, Span: toSourceSpan(input.Span)})
	}
	if activity.Output == "" && activity.Result.Name != "" {
		return Declaration{}, fmt.Errorf("activity %q uses unsupported legacy-only Result; canonical Output is required", activity.Name)
	}
	if activity.Output != "" {
		declaration.Outputs = append(declaration.Outputs, Reference{Name: activity.Output, Span: toSourceSpan(activity.Span)})
	}
	return declaration, nil
}

func toSourceSpan(span syntax.Span) SourceSpan {
	return SourceSpan{
		File:        span.Filename,
		Start:       span.Start.Offset,
		End:         span.End.Offset,
		StartLine:   span.Start.Line,
		StartColumn: span.Start.Column,
		EndLine:     span.End.Line,
		EndColumn:   span.End.Column,
	}
}

func toSemanticSpan(span SourceSpan) semantic.Span {
	return semantic.Span{
		File:  span.File,
		Start: semantic.Position{Offset: span.Start, Line: span.StartLine, Column: span.StartColumn},
		End:   semantic.Position{Offset: span.End, Line: span.EndLine, Column: span.EndColumn},
	}
}
