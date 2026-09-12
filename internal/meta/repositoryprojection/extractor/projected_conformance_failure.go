package extractor

import (
	"path/filepath"
	"strings"
)

const maxProjectedDiagnosticRunes = 4096

// projectedConformanceFailure preserves diagnostics without changing admission.
func projectedConformanceFailure(root, location, phase string, cause error) error {
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
	return failWithDiagnostics("verify-result", "projected-conformance",
		"PROJECTED_CONFORMANCE_FAILED", "KNOWN_CONTRADICTION", "report-counterexample",
		[]string{location, "phase=" + phase, "cause=" + detail})
}
