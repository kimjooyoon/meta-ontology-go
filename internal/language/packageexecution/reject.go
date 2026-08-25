package packageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

func baseReceipt(request Request, sources []SourceEvidence, events []Event) Receipt {
	return Receipt{
		Schema:      ReceiptSchema,
		PackagePath: request.PackagePath,
		Entry:       request.Entry,
		Sources:     nonNilSources(sources),
		Events:      nonNilEvents(events),
		Diagnostics: []Diagnostic{},
	}
}

func reject(request Request, reason, resolution string, diagnostics []Diagnostic, sources []SourceEvidence, events []Event) Receipt {
	receipt := baseReceipt(request, sources, events)
	receipt.Decision = "FAIL_CLOSED"
	receipt.Reason = reason
	receipt.Resolution = resolution
	receipt.Diagnostics = nonNilDiagnostics(diagnostics)
	seal(&receipt)
	return receipt
}

func rejectExecution(receipt Receipt, reason, resolution string, diagnostics []sourceexecution.Diagnostic) Receipt {
	receipt.Decision = "FAIL_CLOSED"
	receipt.Reason = reason
	receipt.Resolution = resolution
	receipt.Diagnostics = make([]Diagnostic, 0, len(diagnostics))
	for _, item := range diagnostics {
		receipt.Diagnostics = append(receipt.Diagnostics, Diagnostic{Stage: item.Stage, Code: item.Code, Message: item.Message})
	}
	seal(&receipt)
	return receipt
}

func nonNilSources(values []SourceEvidence) []SourceEvidence {
	if values == nil {
		return []SourceEvidence{}
	}
	return values
}

func nonNilEvents(values []Event) []Event {
	if values == nil {
		return []Event{}
	}
	return values
}

func nonNilDiagnostics(values []Diagnostic) []Diagnostic {
	if values == nil {
		return []Diagnostic{}
	}
	return values
}
