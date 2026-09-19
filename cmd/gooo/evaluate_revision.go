package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const evaluateRevisionUsage = "usage: gooo evaluate-revision <baseline.gooo> <candidate.gooo> --revision <revision.json> --activity <name> --input <input.json> --out <directory>"

type evaluateRevisionOptions struct {
	baseline, candidate, revision, activity, input, outputDir string
}

func runEvaluateRevision(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseEvaluateRevisionArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, evaluateRevisionUsage)
		return exitUsage
	}
	baselineSource, err := readSource(reader, options.baseline)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	candidateSource, err := readSource(reader, options.candidate)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	revisionData, err := reader.ReadFile(options.revision)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var revision valueexecution.SourceRevision
	if err := json.Unmarshal(revisionData, &revision); err != nil {
		fmt.Fprintf(stderr, "gooo: evaluate-revision: decode revision: %v\n", err)
		return exitFailure
	}
	inputData, err := reader.ReadFile(options.input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	input, err := decodePlanInput(inputData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: evaluate-revision: input: %v\n", err)
		return exitFailure
	}
	evaluation := valueexecution.EvaluateSourceRevision(revision, options.baseline, baselineSource, options.candidate, candidateSource, options.activity, input)
	payload, err := json.MarshalIndent(evaluation, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: evaluate-revision: encode: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeRevisionEvaluation(options.outputDir, payload); err != nil {
		fmt.Fprintf(stderr, "gooo: evaluate-revision: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "source revision evaluation: %s\n", filepath.Join(options.outputDir, "evaluation.json"))
	if evaluation.State == valueexecution.ReplayClosed {
		return exitOK
	}
	return exitFailure
}

func parseEvaluateRevisionArguments(args []string) (evaluateRevisionOptions, error) {
	var options evaluateRevisionOptions
	positionals := make([]string, 0, 2)
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			positionals = append(positionals, args[index])
			continue
		}
		if index+1 >= len(args) || seen[args[index]] {
			return evaluateRevisionOptions{}, fmt.Errorf("%s", evaluateRevisionUsage)
		}
		seen[args[index]] = true
		index++
		switch args[index-1] {
		case "--revision":
			options.revision = args[index]
		case "--activity":
			options.activity = args[index]
		case "--input":
			options.input = args[index]
		case "--out":
			options.outputDir = args[index]
		default:
			return evaluateRevisionOptions{}, fmt.Errorf("%s", evaluateRevisionUsage)
		}
	}
	if len(positionals) != 2 || options.revision == "" || options.activity == "" || options.input == "" || options.outputDir == "" {
		return evaluateRevisionOptions{}, fmt.Errorf("%s", evaluateRevisionUsage)
	}
	options.baseline, options.candidate = positionals[0], positionals[1]
	return options, nil
}

func writeRevisionEvaluation(outputDir string, payload []byte) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	return os.WriteFile(filepath.Join(outputDir, "evaluation.json"), payload, 0o644)
}
