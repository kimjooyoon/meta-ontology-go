package packageexecution

import (
	"path"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

func Execute(request Request) Receipt {
	normalized, issue := normalizeRequest(request)
	if issue != nil {
		return reject(request, issue.Code, "EXACT", []Diagnostic{*issue}, nil, nil)
	}
	parsed, diagnostics, reason := parsePackage(normalized)
	if reason != "" {
		return reject(normalized, reason, "EXACT", diagnostics, parsed.evidence, parsed.events)
	}
	nested := sourceexecution.Execute(sourceexecution.Request{
		Filename: path.Join(normalized.PackagePath, "__package__.gooo"),
		Source:   parsed.source,
		Entry:    normalized.Entry,
	})
	return reduceExecution(normalized, parsed, nested)
}

func reduceExecution(request Request, parsed parsedPackage, nested sourceexecution.Receipt) Receipt {
	receipt := baseReceipt(request, parsed.evidence, parsed.events)
	receipt.Package = parsed.packageName
	receipt.Namespace = parsed.namespace
	receipt.CombinedSourceDigest = digestBytes([]byte(parsed.source))
	receipt.SemanticDigest = nested.SemanticDigest
	receipt.Execution = &nested
	receipt.Effects = Effects{RepositoryWrites: nested.Effects.RepositoryWrites, MutationAuthority: nested.Effects.MutationAuthority}
	appendExecutionEvents(&receipt, nested)
	if receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return rejectExecution(receipt, "PACKAGE_EFFECT_BOUNDARY_VIOLATED", "EXACT", nested.Diagnostics)
	}
	switch nested.Decision {
	case "PASS":
		receipt.Decision = "PASS"
		receipt.Reason = "PACKAGE_EXECUTED"
		receipt.Resolution = "EXACT"
		seal(&receipt)
		return receipt
	case "FAIL_CLOSED":
		return rejectExecution(receipt, "PACKAGE_EXECUTION_REJECTED", "EXACT", nested.Diagnostics)
	default:
		return rejectExecution(receipt, "PACKAGE_EXECUTION_DECISION_UNKNOWN", "LOWER_RESOLUTION", nested.Diagnostics)
	}
}

func appendExecutionEvents(receipt *Receipt, nested sourceexecution.Receipt) {
	for _, event := range nested.Events {
		receipt.Events = append(receipt.Events, Event{Sequence: len(receipt.Events) + 1, Kind: event.Kind, Subject: event.Subject})
	}
}
