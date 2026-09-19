package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const compareAcceptedRevisionUsage = "usage: gooo compare-accepted-revision <baseline.gooo> <candidate.gooo> --revision <revision.json> --evaluation <evaluation.json> --accepted <accepted.json> --activity <name> --input <input.json>"

type compareAcceptedRevisionOptions struct {
	baseline, candidate, revision, evaluation, accepted, activity, input string
}

func runCompareAcceptedRevision(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseCompareAcceptedRevisionArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, compareAcceptedRevisionUsage)
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
	var revision valueexecution.SourceRevision
	if err := decodeJSONFile(reader, options.revision, &revision); err != nil {
		fmt.Fprintf(stderr, "gooo: compare-accepted-revision: decode revision: %v\n", err)
		return exitFailure
	}
	var evaluation valueexecution.SourceRevisionEvaluation
	if err := decodeJSONFile(reader, options.evaluation, &evaluation); err != nil {
		fmt.Fprintf(stderr, "gooo: compare-accepted-revision: decode evaluation: %v\n", err)
		return exitFailure
	}
	var accepted valueexecution.AcceptedSourceRevisionExecution
	if err := decodeJSONFile(reader, options.accepted, &accepted); err != nil {
		fmt.Fprintf(stderr, "gooo: compare-accepted-revision: decode accepted execution: %v\n", err)
		return exitFailure
	}
	inputData, err := reader.ReadFile(options.input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	input, err := decodePlanInput(inputData)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: compare-accepted-revision: input: %v\n", err)
		return exitFailure
	}
	comparison := valueexecution.CompareAcceptedRevisionNextRun(valueexecution.AcceptedRevisionNextRunRequest{
		Revision: revision, Evaluation: evaluation, Accepted: accepted,
		BaselineFilename: options.baseline, BaselineSource: baseline,
		CandidateFilename: options.candidate, CandidateSource: candidate,
		Activity: options.activity, Input: input,
	})
	if err := json.NewEncoder(stdout).Encode(comparison); err != nil {
		return exitFailure
	}
	if comparison.State != valueexecution.ReplayClosed {
		return exitFailure
	}
	return exitOK
}

func parseCompareAcceptedRevisionArguments(args []string) (compareAcceptedRevisionOptions, error) {
	var options compareAcceptedRevisionOptions
	positionals := make([]string, 0, 2)
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			positionals = append(positionals, args[index])
			continue
		}
		flag := args[index]
		if seen[flag] || index+1 >= len(args) {
			return compareAcceptedRevisionOptions{}, fmt.Errorf("%s", compareAcceptedRevisionUsage)
		}
		seen[flag] = true
		index++
		switch flag {
		case "--revision":
			options.revision = args[index]
		case "--evaluation":
			options.evaluation = args[index]
		case "--accepted":
			options.accepted = args[index]
		case "--activity":
			options.activity = args[index]
		case "--input":
			options.input = args[index]
		default:
			return compareAcceptedRevisionOptions{}, fmt.Errorf("%s", compareAcceptedRevisionUsage)
		}
	}
	if len(positionals) != 2 || options.revision == "" || options.evaluation == "" || options.accepted == "" || options.activity == "" || options.input == "" {
		return compareAcceptedRevisionOptions{}, fmt.Errorf("%s", compareAcceptedRevisionUsage)
	}
	options.baseline, options.candidate = positionals[0], positionals[1]
	return options, nil
}

func decodeJSONFile(reader SourceReader, filename string, target any) error {
	data, err := reader.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
