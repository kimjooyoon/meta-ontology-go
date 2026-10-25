package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"strings"
)

// The report's owner is retained, not rediscovered from the current checkout.
func validateModuleOrigins(report sourceReport) error {
	owners := make(map[string]string, len(report.Files))
	for _, file := range report.Files {
		if file.ModuleRoot != "" && (file.Language != "go" || !validPath(file.ModuleRoot, false) ||
			!strings.HasPrefix(file.Path, file.ModuleRoot+"/")) {
			return sourceValidationFailure("SOURCE_MODULE_OWNER_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
		}
		owners[file.Path] = file.ModuleRoot
	}
	for _, row := range report.Meta.Indicators {
		if row.ModuleRoot != "" && !validPath(row.ModuleRoot, false) {
			return sourceValidationFailure("SOURCE_MODULE_OWNER_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
		}
		path := row.Subject
		switch row.SubjectKind {
		case "FILE":
		case "FUNCTION":
			subject, err := sourcepolicy.ParseSourceSubject(row.Subject)
			if err != nil {
				return sourceValidationFailure("SOURCE_FUNCTION_ORIGIN_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
			}
			path = subject.Path
		default:
			continue
		}
		if owner, exists := owners[path]; !exists || owner != row.ModuleRoot {
			return sourceValidationFailure("SOURCE_MODULE_OWNER_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
		}
	}
	return nil
}
