package main

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractionplan"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"strings"
)

func extractionLines(data []byte) int { return extractionplan.Lines(data) }

func extractionFailure(logical string, err error) extractionFailureRecord {
	if failure, ok := errors.AsType[projectionextractor.Failure](err); ok {
		decision := "UNKNOWN"
		unknownClass := failure.UnknownClass
		if failure.UnknownClass == "KNOWN_CONTRADICTION" {
			decision, unknownClass = "REFUTED", ""
		}
		return extractionFailureRecord{Logical: logical, BlockerID: stableBlockerID(logical, failure.Diagnostics), Decision: decision, Stage: failure.Stage, Step: failure.Step, Reason: failure.Reason, UnknownClass: unknownClass, NextOperation: failure.NextOperation, BlockedBy: failure.BlockedBy, Diagnostics: failure.Diagnostics}
	}
	return extractionFailureRecord{Logical: logical, Decision: "UNKNOWN", Stage: "apply-extraction", Step: "generic", Reason: "EXTRACTION_FAILED", UnknownClass: "DIRECT_MISSING", NextOperation: "restore-parser-evidence", BlockedBy: []string{}, Diagnostics: []string{}}
}

func stableBlockerID(logical string, diagnostics []string) string {
	for _, diagnostic := range diagnostics {
		if after, ok := strings.CutPrefix(diagnostic, "declaration="); ok {
			return logical + "#" + after
		}
	}
	return ""
}
