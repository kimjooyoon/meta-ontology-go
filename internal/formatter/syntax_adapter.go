package formatter

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SyntaxAdapter adapts the current surface syntax AST to the formatter's
// parser-neutral document. The formatter deliberately supports only the
// entity/activity surface represented by the archived implementation.
type SyntaxAdapter struct{}

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

// FormatSyntax formats a current syntax AST through the recovered formatter.
func FormatSyntax(file *syntax.File) Result {
	return FormatAST(file, SyntaxAdapter{})
}

func syntaxActivityInputs(activity *syntax.ActivityDecl) ([]syntax.NameRef, *Diagnostic) {
	if activity.Inputs != nil && activity.Parameters != nil && !sameSyntaxNames(activity.Inputs, activity.Parameters) {
		return nil, syntaxDiagnostic(CodeInvalidDocument, "activity "+activity.Name+" has conflicting input aliases", activity.Span)
	}
	inputs := activity.Parameters
	if inputs == nil {
		inputs = activity.Inputs
	}
	return inputs, nil
}

func syntaxActivityOutput(activity *syntax.ActivityDecl) (string, *Diagnostic) {
	output := activity.Result.Name
	if output != "" && activity.Output != "" && output != activity.Output {
		return "", syntaxDiagnostic(CodeInvalidDocument, "activity "+activity.Name+" has conflicting result aliases", activity.Span)
	}
	if output == "" {
		output = activity.Output
	}
	return output, nil
}

func sameSyntaxDeclarations(left, right []syntax.Declaration) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameSyntaxNames(left, right []syntax.NameRef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Name != right[index].Name {
			return false
		}
	}
	return true
}

func unsupportedDeclarationDiagnostic(message string, span syntax.Span) Diagnostics {
	return Diagnostics{*syntaxDiagnostic(CodeUnsupportedSyntax, message, span)}
}

func syntaxDiagnostic(code DiagnosticCode, message string, span syntax.Span) *Diagnostic {
	return &Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		Span:     syntaxSpan(span),
	}
}

func syntaxSpan(span syntax.Span) Span {
	return Span{
		Filename: span.Filename,
		Start: Position{
			Offset: span.Start.Offset,
			Line:   span.Start.Line,
			Column: span.Start.Column,
		},
		End: Position{
			Offset: span.End.Offset,
			Line:   span.End.Line,
			Column: span.End.Column,
		},
	}
}
