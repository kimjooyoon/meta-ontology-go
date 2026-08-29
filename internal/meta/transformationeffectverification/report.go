package transformationeffectverification

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func failureReport(cause error) Report {
	report := Report{Schema: verifierSchema, Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION",
		Stage: "verify-bundle", Step: "validate-input", Reason: "UNKNOWN_CAUSE_UNCATALOGED",
		UnknownClass: "UNCATALOGED_CAUSE", NextOperation: "restore-transformation-evidence", BlockedBy: []string{},
		Improvement: "UNKNOWN", OperationOutcome: "UNKNOWN", PromotionAuthorized: false}
	var failure *validationFailure
	if errors.As(cause, &failure) {
		report.Decision, report.Resolution = failure.Decision, failure.Resolution
		report.Stage, report.Step, report.Reason = failure.Stage, failure.Step, failure.Reason
		report.UnknownClass, report.NextOperation, report.BlockedBy = failure.Unknown, failure.Next, failure.Blocked
		report.FieldPath, report.Expected, report.Observed = failure.FieldPath, failure.Expected, failure.Observed
	}
	return report
}

func WriteReport(path string, report Report) error {
	if path == "" {
		return fmt.Errorf("verification report path is required")
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}
