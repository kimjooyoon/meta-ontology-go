package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyCodegenUsage = "usage: gooo body-codegen [--json] [--sample-seed <seed>] --activity <name> <file.gooo>"

func runBodyCodegen(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	jsonMode := false
	activity := ""
	sampleSeed := ""
	sampleSeedSet := false
	filename := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--json":
			jsonMode = true
		case "--activity":
			if index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			activity = args[index+1]
			index++
		case "--sample-seed":
			if sampleSeedSet || index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" || strings.HasPrefix(args[index+1], "-") {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			sampleSeed = args[index+1]
			sampleSeedSet = true
			index++
		default:
			if strings.HasPrefix(args[index], "-") || filename != "" {
				fmt.Fprintln(stderr, bodyCodegenUsage)
				return exitUsage
			}
			filename = args[index]
		}
	}
	if activity == "" || filename == "" {
		fmt.Fprintln(stderr, bodyCodegenUsage)
		return exitUsage
	}
	source, err := reader.ReadFile(filename)
	if err != nil {
		return reportBodyCodegenFailure(jsonMode, filename, activity, nil, err, stdout, stderr)
	}
	result, err := bodycodegen.GenerateWithPlannerAndSampleSeed(context.Background(), filename, source, activity, os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"), sampleSeed)
	if err != nil {
		return reportBodyCodegenFailure(jsonMode, filename, activity, source, err, stdout, stderr)
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(result); err != nil {
			fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
			return exitFailure
		}
		return exitOK
	}
	if _, err := io.WriteString(stdout, result.Source); err != nil {
		fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
		return exitFailure
	}
	return exitOK
}

func reportBodyCodegenFailure(jsonMode bool, filename, activity string, source []byte, cause error, stdout, stderr io.Writer) int {
	if jsonMode {
		completeness := bodycodegen.FailureCompletenessReceipt(activity, source, cause.Error())
		payload := struct {
			Schema              string                           `json:"schema"`
			Decision            string                           `json:"decision"`
			Source              string                           `json:"source"`
			PlanSHA256          string                           `json:"plan_sha256"`
			CompilerSourceSHA   string                           `json:"compiler_source_sha"`
			RepositoryWrites    int                              `json:"repository_writes"`
			Error               string                           `json:"error"`
			CompletenessReceipt *bodycodegen.CompletenessReceipt `json:"completeness_receipt"`
		}{
			Schema: "gooo/body-codegen-report/v3", Decision: "FAIL_CLOSED",
			PlanSHA256:        stringScopeValue(completeness.Scope, "plan_sha256"),
			CompilerSourceSHA: stringScopeValue(completeness.Scope, "compiler_source_sha"),
			RepositoryWrites:  0, Error: cause.Error(), CompletenessReceipt: completeness,
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintf(stderr, "gooo: body-codegen output: %v\n", err)
		}
		return exitFailure
	}
	fmt.Fprintf(stderr, "gooo: %s: body-codegen: %v\n", filename, cause)
	return exitFailure
}

func stringScopeValue(scope map[string]any, key string) string {
	value, _ := scope[key].(string)
	return value
}
