package languageproofartifactverifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func WithValidationFailure(report Report, err error) Report {
	failure := &ValidationFailure{Coordinate: Coordinate{"VALIDATE", "report", "VALIDATION_FAILED"}, Detail: err.Error()}
	var typed *ValidationError
	if errors.As(err, &typed) {
		failure.Coordinate = typed.Coordinate
	}
	report.ValidationFailure = failure
	return report
}

func WriteReport(path string, report Report) error {
	if err := Validate(report); err != nil {
		caseReasons := make([]string, 0, len(report.Cases))
		for _, item := range report.Cases {
			caseReasons = append(caseReasons, item.ID+"="+item.ObservedReason)
		}
		return fmt.Errorf("validate proof-carrying report: %w (conformance=%s/%s authority=%q cases=%d/%d claims=%d/%d discharged=%d open=%d refuted=%d transitions=%d indicators=%d proofs=%d interventions=%d case_reasons=%s)", err,
			report.ConformanceDecision, report.ConformanceResolution, report.ArtifactUseAuthority,
			report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.Summary.ClaimInstances, report.Summary.ClaimTemplates,
			report.Summary.CaseDischargedClaims, report.Summary.CaseOpenClaims, report.Summary.CaseRefutedClaims,
			len(report.Transitions), len(report.Indicators), len(report.Proofs), len(report.Interventions), strings.Join(caseReasons, ","))
	}
	return writeCanonicalReport(path, report)
}

// WritePreliminaryReport validates and seals the exact pre-consumer
// observation. Preliminary output is intentionally a separate API so the
// final writer cannot accidentally accept or emit a report without a receipt.
func WritePreliminaryReport(path string, report Report) error {
	if err := ValidatePreliminary(report); err != nil {
		return fmt.Errorf("validate preliminary proof-carrying report: %w", err)
	}
	return writeCanonicalReport(path, report)
}

func writeCanonicalReport(path string, report Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func LoadReport(path string) (Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Report{}, err
	}
	return decodeStrict[Report](raw)
}
