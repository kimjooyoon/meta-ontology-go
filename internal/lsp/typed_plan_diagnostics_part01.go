package lsp

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const typedPlanDiagnosticCode = "semantic.typed-plan"

func validateTypedPlanForLSP(file *syntax.File, support syntax.EntityFieldsSupport) error {
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return err
	}
	if len(document.BindingEdges) == 0 {
		return nil
	}
	_, err = bidir.CompileTypedPlan(document)
	return err
}

func typedPlanDiagnostic(uri, source string, err error) (Diagnostic, error) {
	start, startErr := OffsetToPosition(source, 0)
	if startErr != nil {
		return Diagnostic{}, startErr
	}
	end, endErr := OffsetToPosition(source, len(source))
	if endErr != nil {
		return Diagnostic{}, endErr
	}
	return Diagnostic{
		Range:    Range{Start: start, End: end},
		Severity: DiagnosticError,
		Code:     typedPlanDiagnosticCode,
		Source:   "gooo",
		Message:  fmt.Sprintf("typed runtime plan is invalid: %v", err),
		filename: uri,
		start:    0,
		end:      len(source),
		spanned:  true,
	}, nil
}
