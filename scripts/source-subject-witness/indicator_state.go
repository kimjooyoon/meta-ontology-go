package main

import (
	"encoding/hex"
	"fmt"
)

type sourceValidationError struct {
	Decision     string
	Resolution   string
	Stage        string
	Step         string
	Reason       string
	UnknownClass string
	NextOperation string
	BlockedBy    []string
}

func (err *sourceValidationError) Error() string {
	return fmt.Sprintf("source observation decision=%s resolution=%s stage=%s step=%s reason=%s unknown_class=%s next_operation=%s blocked_by=%v", err.Decision, err.Resolution, err.Stage, err.Step, err.Reason, err.UnknownClass, err.NextOperation, err.BlockedBy)
}

func sourceValidationFailure(reason, unknownClass, nextOperation string) error {
	decision, resolution := "FAIL_CLOSED", "LOWER_RESOLUTION"
	if unknownClass == "KNOWN_CONTRADICTION" {
		decision, resolution, unknownClass = "REFUTED", "INVARIANT_ONLY", ""
	}
	return &sourceValidationError{Decision: decision, Resolution: resolution, Stage: "validate-source", Step: "validate-indicator", Reason: reason, UnknownClass: unknownClass, NextOperation: nextOperation, BlockedBy: []string{}}
}

func validCommitSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validateIndicatorShape(row sourceIndicator) error {
	if row.Subject == "" || row.SubjectKind == "" || row.MetricID == "" || row.Relation == "" || row.ApplicabilityRuleID == "" || row.ApplicabilityReason == "" || row.EvaluationState == "" || row.FailureReason == "" || row.EnforcementEffect == "" || row.MetaOperation == "" || row.Producer == "" || row.Consumer == "" || row.ProofChoice == "" || row.Value < 0 || row.Limit < 0 {
		return sourceValidationFailure("SOURCE_INDICATOR_INCOMPLETE", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	if row.Applicability != "APPLICABLE" && row.Applicability != "NOT_APPLICABLE" {
		return sourceValidationFailure("SOURCE_INDICATOR_APPLICABILITY_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	if row.Decision != "PASS" && row.Decision != "FAIL_CLOSED" && row.Decision != "NOT_APPLICABLE" {
		return sourceValidationFailure("SOURCE_INDICATOR_DECISION_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	return nil
}

func validateIndicatorState(row sourceIndicator) error {
	if isLineCapMetric(row.MetricID) {
		return validateLineCapIndicator(row)
	}
	if row.Applicability == "NOT_APPLICABLE" {
		if row.Satisfied && row.Decision == "NOT_APPLICABLE" && row.EvaluationState == "EVALUATED" && row.FailureReason == "CATALOG_NOT_APPLICABLE" && !row.Blocking && row.EnforcementEffect == "NO_EFFECT" {
			return nil
		}
		return sourceValidationFailure("SOURCE_NOT_APPLICABLE_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if !row.Satisfied {
		return sourceValidationFailure("SOURCE_INDICATOR_UNEXPECTED_UNSATISFIED", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if row.Decision != "PASS" || row.EvaluationState != "EVALUATED" || row.FailureReason != "NONE" {
		return sourceValidationFailure("SOURCE_INDICATOR_OUTCOME_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}

func isLineCapMetric(metric string) bool {
	return metric == "gooo.metric.source.go-file-lines.v1" || metric == "gooo.metric.source.gooo-file-lines.v1" || metric == functionLinesMetric
}

func validateLineCapIndicator(row sourceIndicator) error {
	kind, family, operation, consumer := lineCapContract(row.MetricID)
	valid := row.Applicability == "APPLICABLE" && row.SubjectKind == kind && row.Family == family && row.ApplicabilityRuleID == defaultApplicabilityRule && row.ApplicabilityReason == "CATALOG_APPLICABLE" && row.Relation == "less_or_equal" && row.Role == "DRIVER" && !row.Blocking && row.Producer == "linecaps.Analyze" && row.Consumer == consumer && row.MetaOperation == operation && row.ProofChoice == "foundation" && row.EvaluationState == "EVALUATED" && row.EnforcementEffect == "NO_EFFECT" && row.Satisfied == (row.Value <= row.Limit)
	if !valid {
		return sourceValidationFailure("SOURCE_LINE_CAP_DRIVER_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	if row.Satisfied {
		if row.Decision != "PASS" || row.FailureReason != "NONE" || row.FailureCode != "" {
			return sourceValidationFailure("SOURCE_LINE_CAP_PASS_OUTCOME_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
		}
		return nil
	}
	if row.Value <= row.Limit || row.Decision != "FAIL_CLOSED" || row.FailureReason != "PREDICATE_FALSE" || row.FailureCode != row.MetricID+"#predicate-false" {
		return sourceValidationFailure("SOURCE_LINE_CAP_UNSATISFIED_OUTCOME_CONTRADICTION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return nil
}

func lineCapContract(metric string) (string, string, string, string) {
	switch metric {
	case "gooo.metric.source.function-lines.v1":
		return "FUNCTION", "duplication", "extract-function", "function-extractor"
	case "gooo.metric.source.gooo-file-lines.v1":
		return "FILE", "volume", "split-gooo-sections", "source-splitter"
	default:
		return "FILE", "volume", "split-go-declarations", "source-splitter"
	}
}
