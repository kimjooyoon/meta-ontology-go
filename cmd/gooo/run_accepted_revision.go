package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const runAcceptedRevisionUsage = "usage: gooo run-accepted-revision <baseline.gooo> <candidate.gooo> --revision <revision.json> --evaluation <evaluation.json> --activity <name> --input <input.json> --accept"

type runAcceptedRevisionOptions struct {
	baseline, candidate, revision, evaluation, activity, input string
	accept                                                     bool
}

func runAcceptedRevision(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseRunAcceptedRevisionArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, runAcceptedRevisionUsage)
		return exitUsage
	}
	baseline, err := readSource(reader, options.baseline)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	candidate, err := readSource(reader, options.candidate)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	revisionData, err := reader.ReadFile(options.revision)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	evaluationData, err := reader.ReadFile(options.evaluation)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	inputData, err := reader.ReadFile(options.input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var revision valueexecution.SourceRevision
	if err := json.Unmarshal(revisionData, &revision); err != nil {
		fmt.Fprintf(stderr, "gooo: run-accepted-revision: decode revision: %v\n", err)
		return exitFailure
	}
	var evaluation valueexecution.SourceRevisionEvaluation
	if err := json.Unmarshal(evaluationData, &evaluation); err != nil {
		fmt.Fprintf(stderr, "gooo: run-accepted-revision: decode evaluation: %v\n", err)
		return exitFailure
	}
	input, err := decodePlanInput(inputData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run-accepted-revision: input: %v\n", err)
		return exitFailure
	}
	result, err := valueexecution.ExecuteAcceptedSourceRevision(valueexecution.AcceptedSourceRevisionRequest{
		Revision: revision, Evaluation: evaluation, BaselineFilename: options.baseline, BaselineSource: baseline,
		CandidateFilename: options.candidate, CandidateSource: candidate, Activity: options.activity, Input: input,
		ExplicitDecision: valueexecution.AcceptedSourceRevisionDecision,
	})
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run-accepted-revision: %v\n", err)
		return exitFailure
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return exitFailure
	}
	return exitOK
}

func parseRunAcceptedRevisionArguments(args []string) (runAcceptedRevisionOptions, error) {
	var options runAcceptedRevisionOptions
	positionals := make([]string, 0, 2)
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			positionals = append(positionals, args[index])
			continue
		}
		flag := args[index]
		if seen[flag] {
			return runAcceptedRevisionOptions{}, fmt.Errorf("%s", runAcceptedRevisionUsage)
		}
		seen[flag] = true
		if flag == "--accept" {
			options.accept = true
			continue
		}
		if index+1 >= len(args) {
			return runAcceptedRevisionOptions{}, fmt.Errorf("%s", runAcceptedRevisionUsage)
		}
		index++
		switch flag {
		case "--revision":
			options.revision = args[index]
		case "--evaluation":
			options.evaluation = args[index]
		case "--activity":
			options.activity = args[index]
		case "--input":
			options.input = args[index]
		default:
			return runAcceptedRevisionOptions{}, fmt.Errorf("%s", runAcceptedRevisionUsage)
		}
	}
	if len(positionals) != 2 || options.revision == "" || options.evaluation == "" || options.activity == "" || options.input == "" || !options.accept {
		return runAcceptedRevisionOptions{}, fmt.Errorf("%s", runAcceptedRevisionUsage)
	}
	options.baseline, options.candidate = positionals[0], positionals[1]
	return options, nil
}
