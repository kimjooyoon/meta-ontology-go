package formatter

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Adapt converts a current syntax file without mutating it. Alias conflicts,
// latent field carriers, nil declarations, and unknown AST shapes fail closed
// so callers never receive rewritten source for unsupported input.
func (SyntaxAdapter) Adapt(ast any) (*Document, Diagnostics) {
	file, ok := ast.(*syntax.File)
	if !ok || file == nil {
		return nil, Diagnostics{{
			Severity: SeverityError,
			Code:     CodeUnsupportedSyntax,
			Message:  fmt.Sprintf("unsupported syntax AST type %T", ast),
		}}
	}

	if file.Decls != nil && file.Declarations != nil && !sameSyntaxDeclarations(file.Decls, file.Declarations) {
		return nil, Diagnostics{{
			Severity: SeverityError,
			Code:     CodeInvalidDocument,
			Message:  "file declaration aliases conflict",
			Span:     syntaxSpan(file.Span),
		}}
	}

	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	document := &Document{}
	if file.Package != nil {
		document.Package = file.Package.Name
	}
	if file.Namespace != nil {
		document.Namespace = file.Namespace.Name
	}

	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value == nil {
				return nil, unsupportedDeclarationDiagnostic("nil entity declaration", file.Span)
			}
			if value.FieldsPresent || len(value.Fields) != 0 {
				return nil, unsupportedDeclarationDiagnostic("entity fields are not representable by the formatter surface", value.Span)
			}
			document.Declarations = append(document.Declarations, Declaration{
				Kind: EntityDeclaration,
				Name: value.Name,
				ID:   value.ID,
			})
		case *syntax.ActivityDecl:
			if value == nil {
				return nil, unsupportedDeclarationDiagnostic("nil activity declaration", file.Span)
			}
			inputs, diagnostic := syntaxActivityInputs(value)
			if diagnostic != nil {
				return nil, Diagnostics{*diagnostic}
			}
			output, diagnostic := syntaxActivityOutput(value)
			if diagnostic != nil {
				return nil, Diagnostics{*diagnostic}
			}
			declaration := Declaration{
				Kind:   ActivityDeclaration,
				Name:   value.Name,
				Inputs: make([]string, len(inputs)),
				Output: output,
			}
			for index, input := range inputs {
				declaration.Inputs[index] = input.Name
			}
			document.Declarations = append(document.Declarations, declaration)
		default:
			return nil, unsupportedDeclarationDiagnostic(fmt.Sprintf("unsupported declaration type %T", declaration), file.Span)
		}
	}
	return document, nil
}
