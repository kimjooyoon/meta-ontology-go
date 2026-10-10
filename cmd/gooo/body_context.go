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
	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

const bodyContextUsage = "usage: gooo body-context [--plan <plan.json>] --activity <name> " +
	"[--model <model.json>] [--feature-version <version>] [--include-plan] [--value-flow] <file.gooo>"

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
	return runSourceBodyContext(ctx, source, raw, assembly, options, stdout, stderr)
}

type bodyContextArgs struct {
	plan, activity, filename, featureVersion, model string
	includePlan, valueFlow, featureExplicit         bool
}

func (o *bodyContextArgs) set(flag, value string) bool {
	var target *string
	switch flag {
	case "--plan":
		target = &o.plan
	case "--activity":
		target = &o.activity
	case "--model":
		target = &o.model
	case "--feature-version":
		target = &o.featureVersion
	}
	if target == nil || *target != "" || strings.TrimSpace(value) == "" || strings.HasPrefix(value, "-") {
		return false
	}
	*target = value
	if flag == "--feature-version" {
		o.featureExplicit = true
	}
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
		case "--value-flow":
			if o.valueFlow {
				return o, false
			}
			o.valueFlow = true
		case "--plan", "--activity", "--feature-version", "--model":
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
	valid := o.featureVersion == decision.SplitContextIntentFeatureVersion || o.featureVersion == decision.SemanticContextIntentFeatureVersion ||
		o.featureVersion == decision.PositionedIntentFeatureVersion || o.featureVersion == jointdecision.FeatureVersion ||
		o.featureVersion == jointdecision.ThreeFeatureVersion || o.featureVersion == jointdecision.ThreeBagFeatureVersion ||
		o.featureVersion == jointdecision.RecordFieldFeatureVersion || o.featureVersion == jointdecision.RecordSharedFeatureVersion ||
		o.featureVersion == jointdecision.RecordOriginSharedFeatureVersion ||
		o.featureVersion == jointdecision.RecordGraphSharedFeatureVersion ||
		o.featureVersion == decision.ConditionChannelFeatureVersion || o.featureVersion == decision.ConditionBranchFeatureVersion ||
		o.featureVersion == decision.ExecutionFeatureVersion || o.featureVersion == decision.ExecutionFlowFeatureVersion ||
		o.featureVersion == decision.SemanticFlowFeatureVersion
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
