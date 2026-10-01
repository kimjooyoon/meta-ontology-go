package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyContextUsage = "usage: gooo body-context --plan <plan.json> --activity <name> <file.gooo>"

func runBodyContext(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyContextWithContext(ctx, args, reader, stdout, stderr)
}

func runBodyContextWithContext(ctx context.Context, args []string, reader SourceReader, stdout, stderr io.Writer) int {
	plan, activity, filename := "", "", ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--plan", "--activity":
			flag := args[i]
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" || strings.HasPrefix(args[i+1], "-") ||
				flag == "--plan" && plan != "" || flag == "--activity" && activity != "" {
				fmt.Fprintln(stderr, bodyContextUsage)
				return exitUsage
			}
			i++
			if flag == "--plan" {
				plan = args[i]
			} else {
				activity = args[i]
			}
		default:
			if strings.HasPrefix(args[i], "-") || filename != "" {
				fmt.Fprintln(stderr, bodyContextUsage)
				return exitUsage
			}
			filename = args[i]
		}
	}
	if plan == "" || activity == "" || filename == "" {
		fmt.Fprintln(stderr, bodyContextUsage)
		return exitUsage
	}
	source, err := reader.ReadFile(filename)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	raw, err := reader.ReadFile(plan)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	document, err := pathplan.DecodeDocument(raw)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	result, err := bodycodegen.ExportTypedPathContext(ctx, filename, source, activity, document)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	if err = json.NewEncoder(stdout).Encode(result); err != nil {
		fmt.Fprintln(stderr, "context export output failed")
		return exitFailure
	}
	return exitOK
}

func bodyContextFailure(stdout io.Writer, err error) int {
	var failure *bodycodegen.BodyPathError
	output := struct {
		Schema  string                       `json:"schema"`
		Status  string                       `json:"status"`
		Error   string                       `json:"error"`
		Receipt *bodycodegen.BodyPathReceipt `json:"receipt,omitempty"`
	}{Schema: "gooo/compiler-path-input-export/v1", Status: "FAIL_CLOSED", Error: err.Error()}
	if errors.As(err, &failure) {
		output.Receipt = failure.Receipt
	}
	_ = json.NewEncoder(stdout).Encode(output)
	return exitFailure
}
