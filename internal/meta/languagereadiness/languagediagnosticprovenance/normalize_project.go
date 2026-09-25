package languagediagnosticprovenance

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/formatter"
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

func projectDiagnostic(observation Observation, target Span) lsp.Diagnostic {
	severity := lsp.DiagnosticError
	if observation.Severity == formatter.SeverityWarning {
		severity = lsp.DiagnosticWarning
	}
	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lspPosition(target.Start),
			End:   lspPosition(target.End),
		},
		Severity: severity,
		Code:     observation.Code,
		Source:   "gooo-" + strings.ToLower(observation.Stage),
		Message:  observation.Message,
	}
}

func lspPosition(position Position) lsp.Position {
	return lsp.Position{Line: position.Line - 1, Character: position.Column - 1}
}
