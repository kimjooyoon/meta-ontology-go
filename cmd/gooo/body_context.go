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

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyContextUsage = "usage: gooo body-context [--plan <plan.json>] --activity <name> [--feature-version <version>] [--include-plan] <file.gooo>"

type bodyContextOutput struct {
	bodycodegen.TypedPathContextExport
	ExpandedPlan *pathplan.Plan `json:"expanded_plan,omitempty"`
}

func runBodyContext(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyContextWithContext(ctx, args, reader, stdout, stderr)
}

func runBodyContextWithContext(ctx context.Context, args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, ok := parseBodyContextArgs(args)
	if !ok {
		fmt.Fprintln(stderr, bodyContextUsage)
		return exitUsage
	}
	source, err := reader.ReadFile(options.filename)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	var raw []byte
	if options.plan != "" {
		raw, err = reader.ReadFile(options.plan)
		if err != nil {
			return bodyContextFailure(stdout, err)
		}
	}
	assembly, err := bodycodegen.SourceAssembly(ctx, options.filename, source, options.activity)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	if bodycodegen.IsRecordAssembly(assembly) {
		if len(raw) != 0 {
			return bodyContextFailure(stdout, fmt.Errorf("record source assembly owns its plan"))
		}
		result, err := bodycodegen.ExportRecordAssemblyContext(ctx, options.filename, source, options.activity, options.includePlan)
		if err != nil {
			return bodyContextFailure(stdout, err)
		}
		if err = json.NewEncoder(stdout).Encode(result); err != nil {
			return bodyContextFailure(stdout, err)
		}
		return exitOK
	}
	document, err := bodycodegen.DecodeSourcePathDocument(ctx, options.filename, source, options.activity, raw)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	result, err := bodycodegen.ExportTypedPathContextWithFeature(ctx, options.filename, source,
		options.activity, document, options.featureVersion)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	output := bodyContextOutput{TypedPathContextExport: result}
	if options.includePlan {
		output.ExpandedPlan = &document.Plan
	}
	if err = json.NewEncoder(stdout).Encode(output); err != nil {
		fmt.Fprintln(stderr, "context export output failed")
		return exitFailure
	}
	return exitOK
}

type bodyContextArgs struct {
	plan, activity, filename, featureVersion string
	includePlan                              bool
}

func (o *bodyContextArgs) set(flag, value string) bool {
	var target *string
	switch flag {
	case "--plan":
		target = &o.plan
	case "--activity":
		target = &o.activity
	case "--feature-version":
		target = &o.featureVersion
	}
	if target == nil || *target != "" || strings.TrimSpace(value) == "" || strings.HasPrefix(value, "-") {
		return false
	}
	*target = value
	return true
}

func parseBodyContextArgs(args []string) (bodyContextArgs, bool) {
	var o bodyContextArgs
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--include-plan":
			if o.includePlan {
				return o, false
			}
			o.includePlan = true
		case "--plan", "--activity", "--feature-version":
			if i+1 >= len(args) || !o.set(args[i], args[i+1]) {
				return o, false
			}
			i++
		default:
			if strings.HasPrefix(args[i], "-") || o.filename != "" {
				return o, false
			}
			o.filename = args[i]
		}
	}
	if o.featureVersion == "" {
		o.featureVersion = decision.SplitContextIntentFeatureVersion
	}
	valid := o.featureVersion == decision.SplitContextIntentFeatureVersion || o.featureVersion == decision.SemanticContextIntentFeatureVersion
	return o, o.activity != "" && o.filename != "" && valid
}

func bodyContextFailure(stdout io.Writer, err error) int {
	output := struct {
		Schema  string                       `json:"schema"`
		Status  string                       `json:"status"`
		Error   string                       `json:"error"`
		Receipt *bodycodegen.BodyPathReceipt `json:"receipt,omitempty"`
	}{Schema: "gooo/compiler-path-input-export/v1", Status: "FAIL_CLOSED", Error: err.Error()}
	if failure, ok := errors.AsType[*bodycodegen.BodyPathError](err); ok {
		output.Receipt = failure.Receipt
	}
	_ = json.NewEncoder(stdout).Encode(output)
	return exitFailure
}
