package coupling

import (
	"context"
)

func reasonDiagnostic(reason Reason) (string, int, string) {
	switch reason {
	case ReasonAmbiguous:
		return DiagnosticAmbiguous, DiagnosticWarning, "The coupling explanation is ambiguous."
	case ReasonStaleSnapshot:
		return DiagnosticStaleSnapshot, DiagnosticWarning, "The coupling explanation is stale."
	case ReasonUpstreamUnknown:
		return DiagnosticUpstreamUnknown, DiagnosticWarning, "The upstream coupling explanation is unknown."
	case ReasonUpstreamFail:
		return DiagnosticUpstreamFail, DiagnosticError, "The upstream coupling explanation failed closed."
	case ReasonUnregistered:
		return DiagnosticNoBinding, DiagnosticWarning, "The code surface is not registered in the coupling snapshot."
	case ReasonMissing:
		return DiagnosticNoBinding, DiagnosticWarning, "The coupling explanation is missing."
	default:
		return DiagnosticInvalidEnvelope, DiagnosticError, "The coupling explanation is invalid."
	}
}
func failure(outcome Outcome, code string, severity int, message string, span Range) Result {
	return Result{
		Outcome: outcome,
		Diagnostics: []Diagnostic{{
			Range: span, Severity: severity, Code: code, Source: diagnosticSource, Message: message,
		}},
	}
}
func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
func positionInRange(position Position, value Range) bool {
	if position.Line < value.Start.Line || position.Line > value.End.Line {
		return false
	}
	if position.Line == value.Start.Line && position.Character < value.Start.Character {
		return false
	}
	if position.Line == value.End.Line && position.Character >= value.End.Character {
		return false
	}
	return true
}
