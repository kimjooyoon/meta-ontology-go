package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyrefinement"
)

const bodyRefineUsage = "usage: gooo body-refine --source <source.gooo> --activity <name> --feedback-cases <cases.json> --policy <policy.gooo> [--search-policy] --max-attempts <1..64> [--max-rounds <1..8>] [--evaluation-cases <cases.json>] [--model <model.json>] [--go-bin <go>] --out <new-directory>"

type bodyRefineFlags struct {
	source, feedback, evaluation, policy, out string
	options                                   bodyrefinement.Options
}

func parseBodyRefine(args []string, stderr io.Writer) (bodyRefineFlags, error) {
	var values bodyRefineFlags
	f := flag.NewFlagSet("body-refine", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&values.source, "source", "", "original Gooo construction source")
	f.StringVar(&values.feedback, "feedback-cases", "", "adaptive feedback cases")
	f.StringVar(&values.evaluation, "evaluation-cases", "", "separate final evaluation; never given to policy")
	f.StringVar(&values.policy, "policy", "", "Gooo source with one Decide activity")
	f.BoolVar(&values.options.SearchPolicy, "search-policy", false, "extend policy inputs with search observations and source-declared alternatives")
	f.StringVar(&values.out, "out", "", "new output directory")
	f.StringVar(&values.options.Activity, "activity", "", "assembly activity whose budget may increase")
	f.IntVar(&values.options.MaxAttempts, "max-attempts", 0, "maximum attempt budget per round")
	f.IntVar(&values.options.MaxRounds, "max-rounds", 8, "maximum rounds")
	f.StringVar(&values.options.ModelPath, "model", "", "optional local construction model")
	f.StringVar(&values.options.GoBinary, "go-bin", "", "matching Go executable")
	if err := f.Parse(args); err != nil {
		return values, err
	}
	if f.NArg() != 0 || values.source == "" || values.feedback == "" || values.policy == "" || values.out == "" || values.options.Activity == "" {
		return values, fmt.Errorf("%s", bodyRefineUsage)
	}
	return values, nil
}

func runBodyRefine(args []string, stdout, stderr io.Writer) int {
	values, err := parseBodyRefine(args, stderr)
	if err == flag.ErrHelp {
		return exitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return executeBodyRefine(ctx, values, stdout, stderr)
}

func executeBodyRefine(ctx context.Context, values bodyRefineFlags, stdout, stderr io.Writer) int {
	fail := func(err error) int { fmt.Fprintf(stderr, "gooo body-refine: %v\n", err); return exitFailure }
	if _, err := os.Lstat(values.out); !os.IsNotExist(err) {
		return fail(fmt.Errorf("output must be a new directory"))
	}
	source, feedback, options, err := readBodyRefinement(values)
	if err != nil {
		return fail(err)
	}
	result, runErr := bodyrefinement.Run(ctx, values.source, source, feedback, options)
	if err := writeBodyRefinement(values.out, result); err != nil {
		return fail(err)
	}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		return fail(err)
	}
	if runErr != nil {
		return fail(runErr)
	}
	return exitOK
}

func readBodyRefinement(values bodyRefineFlags) ([]byte, bodyexecution.CompositionCases, bodyrefinement.Options, error) {
	options := values.options
	source, err := readBodyExecutionFile(values.source, 128<<10)
	if err != nil {
		return nil, bodyexecution.CompositionCases{}, options, err
	}
	options.PolicySource, err = readBodyExecutionFile(values.policy, 128<<10)
	if err != nil {
		return nil, bodyexecution.CompositionCases{}, options, err
	}
	readCases := func(path string) (bodyexecution.CompositionCases, error) {
		raw, err := readBodyExecutionFile(path, 32<<10)
		if err != nil {
			return bodyexecution.CompositionCases{}, err
		}
		return bodyexecution.DecodeCompositionCases(raw)
	}
	feedback, err := readCases(values.feedback)
	if err == nil && values.evaluation != "" {
		var evaluation bodyexecution.CompositionCases
		evaluation, err = readCases(values.evaluation)
		options.Evaluation = &evaluation
	}
	return source, feedback, options, err
}

func writeBodyRefinement(directory string, result bodyrefinement.Result) error {
	if err := os.Mkdir(directory, 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "refinement.json"), append(raw, '\n'), 0644); err != nil {
		return err
	}
	cases, _ := json.MarshalIndent(result.FeedbackCases, "", "  ")
	for i, round := range result.Rounds {
		output := bodyCompositionOutput{GeneratedNow: true, Composition: round.Composition, Runtime: round.Runtime}
		if err := writeCompositionOutput(filepath.Join(directory, fmt.Sprintf("round-%02d", i+1)), []byte(round.Source), cases, output); err != nil {
			return err
		}
	}
	if result.SelectedRound >= 0 {
		round := result.Rounds[result.SelectedRound]
		return writeCompositionOutput(filepath.Join(directory, "selected"), []byte(round.Source), cases,
			bodyCompositionOutput{GeneratedNow: true, Composition: round.Composition, Runtime: round.Runtime})
	}
	return nil
}
