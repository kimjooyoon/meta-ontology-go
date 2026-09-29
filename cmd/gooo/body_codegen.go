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

const bodyCodegenUsage = "usage: gooo body-codegen [--json] --activity <name> <file.gooo>"

func runBodyCodegen(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	jsonMode := false
	activity := ""
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
		return reportBodyCodegenFailure(jsonMode, filename, err, stdout, stderr)
	}
	result, err := bodycodegen.GenerateWithPlanner(context.Background(), filename, source, activity, os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"))
	if err != nil {
		return reportBodyCodegenFailure(jsonMode, filename, err, stdout, stderr)
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

func reportBodyCodegenFailure(jsonMode bool, filename string, cause error, stdout, stderr io.Writer) int {
	if jsonMode {
		payload := struct {
			Schema           string `json:"schema"`
			Decision         string `json:"decision"`
			Source           string `json:"source"`
			RepositoryWrites int    `json:"repository_writes"`
			Error            string `json:"error"`
		}{Schema: "gooo/body-codegen-report/v2", Decision: "FAIL_CLOSED", RepositoryWrites: 0, Error: cause.Error()}
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
