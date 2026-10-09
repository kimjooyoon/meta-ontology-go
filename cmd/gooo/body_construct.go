package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodyConstructUsage = "usage: gooo body-construct --source <source.gooo> --cases <evaluation.json> " +
	"(--construction-cases <feedback.json> --attempts <1..64> [--entry <activity>] [--model <model.json>] [--fill-model <model.json>] | " +
	"--construction <construction.json>) [--go-bin <go1.27.2>] [--out <new-directory>]"

type bodyConstructOutput struct {
	GeneratedNow bool                            `json:"generated_now"`
	Construction bodyexecution.JointConstruction `json:"construction"`
	Evaluation   bodyexecution.JointEvaluation   `json:"evaluation"`
}

func runBodyConstruct(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	flags, err := parseBodyConstruct(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	return executeBodyConstruct(ctx, flags, stdout, stderr)
}

func parseBodyConstruct(args []string) (map[string]string, error) {
	flags := map[string]string{"--source": "", "--cases": "", "--construction-cases": "", "--attempts": "",
		"--entry": "", "--model": "", "--fill-model": "", "--construction": "", "--go-bin": "", "--out": ""}
	for i := 0; i < len(args); i += 2 {
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			return nil, fmt.Errorf("%s", bodyConstructUsage)
		}
		flags[args[i]] = args[i+1]
	}
	if flags["--source"] == "" || flags["--cases"] == "" {
		return nil, fmt.Errorf("%s", bodyConstructUsage)
	}
	if flags["--construction"] != "" {
		for _, key := range []string{"--construction-cases", "--attempts", "--entry", "--model", "--fill-model"} {
			if flags[key] != "" {
				return nil, fmt.Errorf("saved construction excludes %s", key)
			}
		}
	} else {
		budget, err := strconv.Atoi(flags["--attempts"])
		if err != nil || budget < 1 || budget > 64 || flags["--construction-cases"] == "" {
			return nil, fmt.Errorf("%s", bodyConstructUsage)
		}
	}
	return flags, nil
}

func executeBodyConstruct(ctx context.Context, flags map[string]string, stdout, stderr io.Writer) int {
	fail := func(err error) int { fmt.Fprintf(stderr, "gooo body-construct: %v\n", err); return exitFailure }
	if directory := flags["--out"]; directory != "" {
		if _, err := os.Lstat(directory); !os.IsNotExist(err) {
			return fail(fmt.Errorf("output must be a new directory"))
		}
	}
	source, err := readBodyExecutionFile(flags["--source"], 128<<10)
	if err != nil {
		return fail(err)
	}
	evaluation, err := readJointCases(flags["--cases"])
	if err != nil {
		return fail(err)
	}
	output := bodyConstructOutput{GeneratedNow: flags["--construction"] == ""}
	if output.GeneratedNow {
		feedback, readErr := readJointCases(flags["--construction-cases"])
		if readErr != nil {
			return fail(readErr)
		}
		budget, _ := strconv.Atoi(flags["--attempts"])
		output.Construction, output.Evaluation, err = bodyexecution.ConstructAndEvaluateJoint(ctx, flags["--source"], source,
			feedback, evaluation, bodyexecution.JointOptions{EntryActivity: flags["--entry"], ProgramBudget: budget,
				ModelPath: flags["--model"], FillModelPath: flags["--fill-model"], GoBinary: flags["--go-bin"]})
	} else {
		raw, readErr := readBodyExecutionFile(flags["--construction"], 32<<20)
		if readErr != nil {
			return fail(readErr)
		}
		output.Construction, err = bodyexecution.DecodeJointConstruction(raw)
		if err == nil {
			output.Evaluation, err = bodyexecution.ReplayJointComposition(ctx, flags["--source"], source, output.Construction, evaluation, flags["--go-bin"])
		}
	}
	if directory := flags["--out"]; directory != "" {
		if writeErr := writeJointOutput(directory, source, output); writeErr != nil {
			return fail(writeErr)
		}
	}
	if encodeErr := json.NewEncoder(stdout).Encode(output); encodeErr != nil {
		return fail(encodeErr)
	}
	if err != nil {
		return fail(err)
	}
	return exitOK
}

func readJointCases(path string) (bodyexecution.CompositionCases, error) {
	raw, err := readBodyExecutionFile(path, 32<<10)
	if err != nil {
		return bodyexecution.CompositionCases{}, err
	}
	return bodyexecution.DecodeCompositionCases(raw)
}
