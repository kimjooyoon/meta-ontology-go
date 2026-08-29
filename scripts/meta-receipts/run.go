package main

import (
	"fmt"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type options struct {
	planPath     string
	receiptsPath string
	outputPath   string
}

func run(configuration options) error {
	if !optionsKnown(configuration) {
		return fmt.Errorf("plan and output must be distinct non-empty paths")
	}
	plan := generation.Plan{}
	if err := decodeJSON(configuration.planPath, &plan); err != nil {
		return err
	}
	receipts := []generation.OperationReceipt{}
	failures := []generation.ObservationFailure{}
	if configuration.receiptsPath != "" {
		if err := decodeJSON(configuration.receiptsPath, &receipts); err != nil {
			return err
		}
	} else {
		manifestPath := filepath.Join(filepath.Dir(configuration.planPath), "self-improvement-execution.json")
		manifest := generation.ExecutionManifest{}
		if err := decodeJSON(manifestPath, &manifest); err != nil {
			return fmt.Errorf("read execution manifest: %w", err)
		}
		bundlePath := filepath.Join(filepath.Dir(configuration.planPath), "meta-operation-observations.json")
		bundle := generation.OperationObservationBundle{}
		if err := decodeJSON(bundlePath, &bundle); err != nil {
			return fmt.Errorf("read operation observations: %w", err)
		}
		if err := generation.ValidateObservationBundle(bundle, plan, manifest); err != nil {
			return fmt.Errorf("operation observation binding failed: %w", err)
		}
		receipts = bundle.Receipts
		failures = bundle.Failures
	}
	report := generation.VerifyReceiptsWithFailures(plan, receipts, failures)
	payload, err := generation.EncodeReceiptReport(report)
	if err != nil {
		return fmt.Errorf("encode receipt report: %w", err)
	}
	if err := writeAtomic(configuration.outputPath, payload); err != nil {
		return err
	}
	fmt.Printf(
		"receipt verification: decision=%s reason=%s unknown=%d replay=%s\n",
		report.Decision,
		report.Reason,
		len(report.Unknowns),
		report.ReplayDigest,
	)
	if !receiptOutcomeConformant(plan, report) {
		return fmt.Errorf(
			"receipt verification failed: %s/%s",
			report.Decision,
			report.Reason,
		)
	}
	return nil
}

func receiptOutcomeConformant(plan generation.Plan, report generation.ReceiptReport) bool {
	if report.Decision == generation.ReceiptDecisionFixedPoint ||
		report.Decision == generation.ReceiptDecisionConformant {
		return true
	}
	return report.Decision == generation.ReceiptDecisionRefuted && validMixedRefutation(plan, report)
}

func validMixedRefutation(plan generation.Plan, report generation.ReceiptReport) bool {
	if len(plan.Selected) != 2 || len(report.Receipts) != 1 || len(report.Failures) != 1 ||
		report.PromotionAuthorized || len(report.MissingIndicatorIDs) != 0 ||
		len(report.RejectedIndicatorIDs) != 0 || len(report.Unknowns) != 5 {
		return false
	}
	var split, extract generation.Action
	for _, action := range plan.Selected {
		switch action.Operation {
		case sourcepolicy.OperationSplitGo:
			split = action
		case sourcepolicy.OperationExtractFunction:
			extract = action
		default:
			return false
		}
	}
	if split.IndicatorID == "" || extract.IndicatorID == "" ||
		len(split.RequiredIndicatorIDs) != 6 || len(extract.RequiredIndicatorIDs) != 5 {
		return false
	}
	if !validClosedReceipt(report.Receipts[0], split) {
		return false
	}
	failure := report.Failures[0]
	if failure.ActionIndicatorID != extract.IndicatorID || failure.Decision != "REFUTED" ||
		failure.Stage != "derive-recipe" || failure.Step != "select-declaration" ||
		failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" || failure.NextOperation != "report-counterexample" ||
		failure.BlockedBy == nil || !validRefutedIndicatorLinks(failure, extract) {
		return false
	}
	if len(report.UnknownIndicatorIDs) != 1 ||
		report.UnknownIndicatorIDs[0] != extract.IndicatorID+"::dependency:"+failure.Reason {
		return false
	}
	return validDependencyUnknowns(report.Unknowns, failure, extract)
}

func validClosedReceipt(receipt generation.OperationReceipt, action generation.Action) bool {
	if receipt.ActionIndicatorID != action.IndicatorID || receipt.Operation != action.Operation ||
		len(receipt.Indicators) != len(action.RequiredIndicatorIDs) {
		return false
	}
	allowed := make(map[string]bool, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		allowed[identifier] = true
	}
	for _, indicator := range receipt.Indicators {
		if !allowed[indicator.ID] || indicator.Verdict != generation.IndicatorVerdictPass {
			return false
		}
		delete(allowed, indicator.ID)
	}
	return len(allowed) == 0
}

func validRefutedIndicatorLinks(failure generation.ObservationFailure, action generation.Action) bool {
	if len(failure.FailureEvidence) != len(action.RequiredIndicatorIDs) {
		return false
	}
	allowed := make(map[string]bool, len(action.RequiredIndicatorIDs))
	var counterexample string
	for _, identifier := range action.RequiredIndicatorIDs {
		allowed[identifier] = true
	}
	for _, evidence := range failure.FailureEvidence {
		if !allowed[evidence.IndicatorID] || evidence.Decision != "UNKNOWN" ||
			evidence.Observed != 0 || evidence.Expected != 1 || evidence.Counterexample == "" {
			return false
		}
		if counterexample == "" {
			counterexample = evidence.Counterexample
		} else if counterexample != evidence.Counterexample {
			return false
		}
		delete(allowed, evidence.IndicatorID)
	}
	return len(allowed) == 0
}

func validDependencyUnknowns(unknowns []generation.ReceiptUnknown, failure generation.ObservationFailure, action generation.Action) bool {
	allowed := make(map[string]bool, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		allowed[identifier] = true
	}
	for _, unknown := range unknowns {
		if unknown.ActionIndicatorID != action.IndicatorID || !allowed[unknown.RequiredIndicatorID] ||
			unknown.Stage != failure.Stage || unknown.Step != failure.Step ||
			unknown.Reason != generation.ReceiptReason(failure.Reason) ||
			unknown.UnknownClass != generation.ReceiptUnknownClassDependencyBlocked ||
			unknown.NextOperation != failure.NextOperation || len(unknown.BlockedBy) == 0 {
			return false
		}
		delete(allowed, unknown.RequiredIndicatorID)
	}
	return len(allowed) == 0
}

func optionsKnown(configuration options) bool {
	if configuration.planPath == "" || configuration.outputPath == "" {
		return false
	}
	output := filepath.Clean(configuration.outputPath)
	if output == filepath.Clean(configuration.planPath) {
		return false
	}
	return configuration.receiptsPath == "" ||
		output != filepath.Clean(configuration.receiptsPath)
}
