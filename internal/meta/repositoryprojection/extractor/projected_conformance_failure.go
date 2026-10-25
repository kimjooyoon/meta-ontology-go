package extractor

import (
	"path/filepath"
	"strings"
)

const maxProjectedDiagnosticRunes = 4096

// projectedConformanceFailure preserves diagnostics without changing admission.
func projectedConformanceFailure(root, location, phase string, cause error) error {
	return failWithDiagnostics("verify-result", "projected-conformance",
		"PROJECTED_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample",
		projectedFailureDiagnostics(root, location, phase, cause))
}

// Missing compilation observations cannot refute a source transformation.
// This remains unresolved work; it does not satisfy the conformance obligation.
func projectedDependencyOriginFailure(root, location string, cause error) error {
	return failWithDiagnostics("verify-result", "compilation-origin",
		"COMPILATION_ORIGIN_UNAVAILABLE", "DIRECT_MISSING", "restore-module-compilation-origin",
		projectedFailureDiagnostics(root, location, "dependency-origin", cause))
}

func projectedFailureDiagnostics(root, location, phase string, cause error) []string {
	detail := filepath.ToSlash(cause.Error())
	workspace := strings.TrimSuffix(filepath.ToSlash(filepath.Clean(root)), "/")
	if workspace != "" && workspace != "." {
		location = strings.ReplaceAll(filepath.ToSlash(location), workspace+"/", "<workspace>/")
		detail = strings.ReplaceAll(detail, workspace+"/", "<workspace>/")
	}
	runes := []rune(detail)
	if len(runes) > maxProjectedDiagnosticRunes {
		detail = string(runes[:maxProjectedDiagnosticRunes]) + "...[truncated]"
	}
	return []string{location, "phase=" + phase, "cause=" + detail}
}
