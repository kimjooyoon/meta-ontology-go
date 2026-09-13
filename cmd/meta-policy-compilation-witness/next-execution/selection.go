package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

type goalSelectionOptions struct {
	goal        *string
	digest      *string
	materialize *bool
}

func bindGoalSelectionFlags(flags *flag.FlagSet) goalSelectionOptions {
	return goalSelectionOptions{
		goal: flags.String("goal", "", "separate frozen Gooo target policy"),
		digest: flags.String("goal-digest", "", "expected SHA-256 of the original frozen goal bytes"),
		materialize: flags.Bool("materialize-selection", false, "write selected Gooo to a new private temporary directory"),
	}
}

func (options goalSelectionOptions) enabled() bool {
	return *options.goal != "" || *options.digest != "" || *options.materialize
}

func (options goalSelectionOptions) valid() bool {
	return !options.enabled() || *options.goal != "" && *options.digest != ""
}

func (options goalSelectionOptions) run(ctx context.Context, filename string, source, predecessor, request []byte, pkg, namespace string, stdout, stderr io.Writer) int {
	goal, err := readBounded(*options.goal, 4<<20)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report := policycompilation.SelectNextPolicyForGoal(ctx, filename, source, predecessor, request, pkg, namespace, goal, *options.digest)
	if report.Decision == "SELECTED_FOR_FROZEN_GOOO_GOAL" && *options.materialize {
		if err := materializeGoalSelection(&report); err != nil {
			report.Decision, report.Reason = "UNKNOWN", "SELECTED_SOURCE_MATERIALIZATION_UNAVAILABLE"
			report.Diagnostic = err.Error()
			report.Pending = &policycompilation.PolicyRevisionPending{
				State: "UNKNOWN", Stage: "OUTPUT", Step: "MATERIALIZE_SELECTED_GOOO",
				Reason: report.Reason, UnknownClass: "DIRECT_MISSING",
				NextOperation: "REPEAT_SELECTION_IN_CALLER_OWNED_TEMP_OUTPUT", BlockedBy: []string{},
			}
		}
	}
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	switch report.Decision {
	case "SELECTED_FOR_FROZEN_GOOO_GOAL":
		return 0
	case "REFUTED":
		return 1
	default:
		return 2
	}
}

func materializeGoalSelection(report *policycompilation.GoalSelectionReport) error {
	directory, err := os.MkdirTemp("", "gooo-selected-policy-")
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "selected-policy.gooo")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err == nil {
		_, err = file.WriteString(report.SelectedSource)
		if closeError := file.Close(); err == nil {
			err = closeError
		}
	}
	if err != nil {
		os.RemoveAll(directory)
		return err
	}
	report.MaterializedSource = path
	return nil
}
