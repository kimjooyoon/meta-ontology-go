package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

func finishDiscoveryReport(reader SourceReader, source []byte, report capabilityDiscoveryReport,
	runtimeFlags map[string]string, jsonMode bool, stdout, stderr io.Writer) int {
	execution, runErr := executeDiscoveryGeneration(reader, report.SourcePath, source, report.Generation,
		runtimeFlags["--execute-cases"], runtimeFlags["--go-bin"])
	if runErr != nil && execution == nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, report.SourcePath, "RUNTIME_INPUT_INVALID", runErr.Error())
	}
	if execution != nil {
		attachDiscoveryRuntime(report.Receipt, execution)
		report.Runtime = &execution.Result
	}
	if err := completeness.Validate(report.Receipt); err != nil {
		return reportDiscoverFailure(jsonMode, stdout, stderr, report.SourcePath, "COMPLETENESS_RECEIPT_INVALID", err.Error())
	}
	report.Decision = report.Receipt.Decision
	return emitDiscoveryReport(report, jsonMode, stdout, stderr)
}

func emitDiscoveryReport(report capabilityDiscoveryReport, jsonMode bool, stdout, stderr io.Writer) int {
	exitCode := exitOK
	if report.Decision == "FAIL_CLOSED" {
		exitCode = exitFailure
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return exitFailure
		}
		return exitCode
	}
	if _, err := fmt.Fprintf(stdout, "capability discovery: %s (%s)\ncompleteness: %s, first unresolved: %s\nsource: %s\n",
		report.Query.Response.Status, report.Query.Intent, report.Decision, report.Receipt.FirstUnresolved.ID, report.SourcePath); err != nil {
		return exitFailure
	}
	for _, dimension := range report.Receipt.Dimensions {
		if report.Generation != nil && dimension.ID == "generation_coverage" {
			if _, err := fmt.Fprintf(stdout, "generation replay: %s\ngeneration coverage: %s %d/%d\n",
				report.Generation.ActivityID, dimension.Status, dimension.Numerator, dimension.Denominator); err != nil {
				return exitFailure
			}
		}
		if report.Runtime != nil && dimension.ID == "real_use_case_coverage" {
			if _, err := fmt.Fprintf(stdout, "runtime: %s\nindependent input expectations: %s %d/%d\n",
				report.Runtime.Observation.Stage, dimension.Status, dimension.Numerator, dimension.Denominator); err != nil {
				return exitFailure
			}
		}
	}
	if report.Runtime != nil && report.Runtime.Observation.Failure != "" {
		fmt.Fprintln(stderr, report.Runtime.Observation.Failure)
	}
	for _, capability := range report.Query.Response.Capabilities {
		if _, err := fmt.Fprintf(stdout, "- %s [%s]: %s\n", capability.ID, capability.State, capability.Description); err != nil {
			return exitFailure
		}
	}
	return exitCode
}
