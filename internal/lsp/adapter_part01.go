package lsp

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type loweredSymbolKey struct {
	start int
	end   int
	kind  semantic.Kind
	name  string
}

// adaptSyntaxResult is retained as a small, context-free test seam. The live
// SyntaxParser path uses adaptSyntaxResultContext so lowering participates in
// request cancellation.
func adaptSyntaxResult(uri, source string, file *syntax.File, diagnostics syntax.Diagnostics) (ParseResult, error) {
	return adaptSyntaxResultContext(context.Background(), uri, source, file, diagnostics)
}
func adaptSyntaxResultContext(ctx context.Context, uri, source string, file *syntax.File, diagnostics syntax.Diagnostics) (ParseResult, error) {
	return adaptSyntaxResultContextWithSupport(ctx, uri, source, file, diagnostics, syntax.CurrentEntityFieldsSupport())
}
func adaptSyntaxResultContextWithSupport(ctx context.Context, uri, source string, file *syntax.File, diagnostics syntax.Diagnostics, support syntax.EntityFieldsSupport) (ParseResult, error) {
	result := ParseResult{File: file}
	for _, diagnostic := range diagnostics.SortBySpan() {
		mapped, err := syntaxDiagnostic(source, diagnostic)
		if err != nil {
			return ParseResult{}, err
		}
		result.Diagnostics = append(result.Diagnostics, mapped)
	}

	result.semanticChecked = file != nil && (file.Package != nil || file.Namespace != nil || len(syntaxDeclarations(file)) > 0)
	ids := make(map[loweredSymbolKey]string)
	names := make(map[string]string)
	if result.semanticChecked && !diagnostics.HasErrors() && file.Package != nil && file.Namespace != nil {
		ir, err := bidir.LowerContextWithEntityFieldsSupport(ctx, canonicalSyntaxFile(file), support)
		if err != nil {
			if errors.Is(err, bidir.ErrLowerCanceled) {
				return ParseResult{}, err
			}
			result.Diagnostics = append(result.Diagnostics, semanticDiagnostic(uri, source, file, err))
		} else if typedPlanErr := validateTypedPlanForLSP(canonicalSyntaxFile(file), support); typedPlanErr != nil {
			diagnostic, diagnosticErr := typedPlanDiagnostic(uri, source, typedPlanErr)
			if diagnosticErr != nil {
				return ParseResult{}, diagnosticErr
			}
			result.Diagnostics = append(result.Diagnostics, diagnostic)
		} else {
			result.semanticValid = true
			result.semanticDigest = ir.StableHash()
			ids, names = loweredIdentities(ir)
		}
	}

	if result.semanticValid {
		seedFallbackLSPIdentities(file, ids, names)
	}

	if file != nil {
		if err := appendHeaderSymbols(&result, source, file); err != nil {
			return ParseResult{}, err
		}
		for _, declaration := range syntaxDeclarations(file) {
			if err := appendDeclaration(&result, source, declaration, ids, names); err != nil {
				return ParseResult{}, err
			}
		}
	}
	return normalizeParseResult(uri, source, result), nil
}
func normalizeParseResult(uri, source string, result ParseResult) ParseResult {
	result.Diagnostics = canonicalDiagnosticOrder(uri, source, result.Diagnostics)
	return result
}

func seedFallbackLSPIdentities(file *syntax.File, ids map[loweredSymbolKey]string, names map[string]string) {
	if file == nil || file.Namespace == nil {
		return
	}
	namespace := strings.TrimSpace(file.Namespace.Name)
	if namespace == "" {
		return
	}
	for _, declaration := range syntaxDeclarations(file) {
		var name string
		var span syntax.Span
		var kind semantic.Kind
		var explicit string
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			name, span, kind, explicit = value.Name, value.Span, semantic.Entity, value.ID
		case *syntax.ActivityDecl:
			name, span, kind = value.Name, value.Span, semantic.Activity
		default:
			continue
		}
		id := explicit
		if id == "" {
			id = fallbackLSPIdentity(namespace, kind, name)
		}
		if id == "" {
			continue
		}
		key := loweredSymbolKey{start: span.Start.Offset, end: span.End.Offset, kind: kind, name: name}
		if ids[key] == "" {
			ids[key] = id
		}
		if names[name] == "" {
			names[name] = id
		}
	}
}

func fallbackLSPIdentity(namespace string, kind semantic.Kind, name string) string {
	segment := ""
	switch kind {
	case semantic.Entity:
		segment = "entity"
	case semantic.Activity:
		segment = "activity"
	default:
		return ""
	}
	slug := fallbackLPSSlug(name)
	if slug == "" {
		return ""
	}
	namespace = strings.ReplaceAll(namespace, "_", "-")
	return namespace + "://" + segment + "/" + slug
}

func fallbackLPSSlug(value string) string {
	var builder strings.Builder
	lastDash := false
	var previous rune
	for _, current := range value {
		if unicode.IsUpper(current) && unicode.IsLower(previous) && builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			builder.WriteRune(unicode.ToLower(current))
			lastDash = false
		} else if builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
		previous = current
	}
	return strings.Trim(builder.String(), "-")
}
