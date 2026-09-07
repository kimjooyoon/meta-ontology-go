package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const runSourceUsage = "usage: gooo run [--json] --entry <activity> [--input <input.json> | --record-input <record.json>] <file.gooo|package-directory>"

func runSource(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if handled, code := maybeRunSourcePackage(args, stdout, stderr); handled {
		return code
	}
	args, jsonMode := parseJSONFlag(args)
	options, err := parseRunSourceArguments(args)
	if err != nil {
		return reportUsage(jsonMode, stdout, stderr, "run", runSourceUsage)
	}
	source, err := readSource(reader, options.filename)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "run", options.filename,
			"read", err.Error(), sourceexecutionSpan())
	}
	if options.input != "" {
		if options.record {
			return runSourceRecordPlan(options, source, reader, jsonMode, stdout, stderr)
		}
		return runSourceValuePlan(options, source, reader, jsonMode, stdout, stderr)
	}
	receipt := sourceexecution.Execute(sourceexecution.Request{
		Filename: options.filename, Source: string(source), Entry: options.entry,
	})
	payload, err := sourceexecution.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run receipt: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		if _, err := stdout.Write(payload); err != nil {
			return exitFailure
		}
	} else if receipt.Decision == "PASS" {
		fmt.Fprintf(stdout, "executed: %s.%s(%s) -> %s digest=%s\n", receipt.Entry.Package,
			receipt.Entry.Activity, inputNames(receipt.Entry.Inputs), receipt.Entry.Output.Name, receipt.Digest)
	} else {
		diagnostic := receipt.Diagnostics[0]
		fmt.Fprintf(stderr, "%s: %s: %s\n", options.filename, diagnostic.Code, diagnostic.Message)
	}
	if receipt.Decision == "PASS" {
		return exitOK
	}
	return exitFailure
}

func runSourceValuePlan(options runSourceOptions, source []byte, reader SourceReader, jsonMode bool, stdout, stderr io.Writer) int {
	input, err := reader.ReadFile(options.input)
	if err != nil {
		return reportPlanFailure(jsonMode, stdout, stderr, options.filename, valueexecution.Execution{}, valueexecution.Failure{
			Code: valueexecution.ReasonSourceReadFailed, Stage: "INPUT", Step: "read-plan-input", Detail: err.Error(),
		})
	}
	plan, err := valueexecution.CompilePlan(options.filename, source)
	if err != nil {
		return reportPlanFailure(jsonMode, stdout, stderr, options.filename, valueexecution.Execution{}, err)
	}
	rootInput, err := decodePlanInput(input)
	if err != nil {
		return reportPlanFailure(jsonMode, stdout, stderr, options.filename, valueexecution.Execution{}, err)
	}
	execution, err := plan.Execute(map[string]int64{options.entry: rootInput})
	if err != nil {
		return reportPlanFailure(jsonMode, stdout, stderr, options.filename, execution, err)
	}
	payload := struct {
		Schema              string                   `json:"schema"`
		Decision            string                   `json:"decision"`
		SourcePath          string                   `json:"source_path"`
		SourceDigest        string                   `json:"source_digest"`
		SemanticFingerprint string                   `json:"semantic_fingerprint"`
		Entry               string                   `json:"entry"`
		Execution           valueexecution.Execution `json:"execution"`
	}{
		Schema: "gooo/value-execution-plan/v1", Decision: "PASS", SourcePath: options.filename,
		SourceDigest: plan.SourceDigest, SemanticFingerprint: plan.SemanticFingerprint, Entry: options.entry, Execution: execution,
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(payload); err != nil {
			return reportPlanFailure(jsonMode, stdout, stderr, options.filename, execution, err)
		}
	} else {
		fmt.Fprintf(stdout, "executed value plan: entry=%s activities=%d applies=%d deliveries=%d\n", options.entry, len(execution.Activities), execution.ApplyCalls, execution.Deliveries)
	}
	return exitOK
}

func decodePlanInput(raw []byte) (int64, error) {
	const detail = "plan input must be an integer or {\"value\": integer}"
	var value int64
	if err := json.Unmarshal(raw, &value); err == nil {
		if strings.TrimSpace(string(raw)) == "null" {
			return 0, valueexecution.Failure{Code: valueexecution.ReasonExternalInputUnexpected, Stage: "INPUT", Step: "decode-plan-input", Detail: detail}
		}
		return value, nil
	}
	var envelope struct {
		Value *int64 `json:"value"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil || envelope.Value == nil {
		return 0, valueexecution.Failure{Code: valueexecution.ReasonExternalInputUnexpected, Stage: "INPUT", Step: "decode-plan-input", Detail: detail}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return 0, valueexecution.Failure{Code: valueexecution.ReasonExternalInputUnexpected, Stage: "INPUT", Step: "decode-plan-input", Detail: detail}
	}
	return *envelope.Value, nil
}

func reportPlanFailure(jsonMode bool, stdout, stderr io.Writer, filename string, execution valueexecution.Execution, err error) int {
	if jsonMode {
		failure, ok := valueexecution.FailureOf(err)
		if !ok {
			failure = valueexecution.Failure{Code: valueexecution.Reason(err), Stage: "EXECUTE", Step: "plan-execution", Detail: err.Error()}
		}
		_ = json.NewEncoder(stdout).Encode(struct {
			Schema     string                   `json:"schema"`
			Decision   string                   `json:"decision"`
			SourcePath string                   `json:"source_path"`
			Reason     string                   `json:"reason"`
			Failure    valueexecution.Failure   `json:"failure"`
			Execution  valueexecution.Execution `json:"execution"`
			Error      string                   `json:"error"`
		}{
			Schema: "gooo/value-execution-plan/v1", Decision: "FAIL_CLOSED", SourcePath: filename,
			Reason: valueexecution.Reason(err), Failure: failure, Execution: execution, Error: err.Error(),
		})
	} else {
		fmt.Fprintf(stderr, "%s: value plan: %v\n", filename, err)
	}
	return exitFailure
}

func inputNames(bindings []sourceexecution.Binding) string {
	names := make([]string, len(bindings))
	for index, binding := range bindings {
		names[index] = binding.Name
	}
	return strings.Join(names, ", ")
}
