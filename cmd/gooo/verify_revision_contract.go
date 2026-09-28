package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const verifyRevisionContractUsage = "usage: gooo verify-revision-contract <baseline.gooo> <candidate.gooo> --revision <revision.json> --evaluation <evaluation.json> --activity <name> --inputs <inputs.json> --out <directory>"

type verifyRevisionContractOptions struct {
	baseline, candidate, revision, evaluation, activity, inputs, expectedOutputs, outputDir string
}

func runVerifyRevisionContract(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseVerifyRevisionContractArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, verifyRevisionContractUsage)
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
	evaluationData, err := reader.ReadFile(options.evaluation)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	inputsData, err := reader.ReadFile(options.inputs)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var expectedOutputs map[int64]int64
	if options.expectedOutputs != "" {
		expectedOutputsData, readErr := reader.ReadFile(options.expectedOutputs)
		if readErr != nil {
			fmt.Fprintln(stderr, readErr)
			return exitFailure
		}
		if err := json.Unmarshal(expectedOutputsData, &expectedOutputs); err != nil || len(expectedOutputs) == 0 {
			fmt.Fprintln(stderr, "gooo: verify-revision-contract: expected outputs must be a non-empty JSON object")
			return exitFailure
		}
	}

	var revision valueexecution.SourceRevision
	if err := json.Unmarshal(revisionData, &revision); err != nil {
		fmt.Fprintf(stderr, "gooo: verify-revision-contract: decode revision: %v\n", err)
		return exitFailure
	}
	var evaluation valueexecution.SourceRevisionEvaluation
	if err := json.Unmarshal(evaluationData, &evaluation); err != nil {
		fmt.Fprintf(stderr, "gooo: verify-revision-contract: decode evaluation: %v\n", err)
		return exitFailure
	}
	var inputs []int64
	if err := json.Unmarshal(inputsData, &inputs); err != nil || len(inputs) == 0 {
		fmt.Fprintln(stderr, "gooo: verify-revision-contract: inputs must be a non-empty JSON integer array")
		return exitFailure
	}
	result := valueexecution.VerifySourceRevisionContract(
		revision, evaluation, options.baseline, baselineSource, options.candidate, candidateSource,
		options.activity, valueexecution.SourceRevisionContract{Scope: valueexecution.SourceRevisionContractScope, Inputs: inputs, ExpectedOutputs: expectedOutputs},
	)
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: verify-revision-contract: encode: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeRevisionEvaluation(options.outputDir, payload); err != nil {
		fmt.Fprintf(stderr, "gooo: verify-revision-contract: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "source revision contract evaluation: %s\n", filepath.Join(options.outputDir, "evaluation.json"))
	if result.ContractPreservation && result.Accepted {
		return exitOK
	}
	return exitFailure
}

func parseVerifyRevisionContractArguments(args []string) (verifyRevisionContractOptions, error) {
	var options verifyRevisionContractOptions
	positionals := make([]string, 0, 2)
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			positionals = append(positionals, args[index])
			continue
		}
		flag := args[index]
		if index+1 >= len(args) || seen[flag] {
			return verifyRevisionContractOptions{}, fmt.Errorf("%s", verifyRevisionContractUsage)
		}
		seen[flag] = true
		index++
		switch flag {
		case "--revision":
			options.revision = args[index]
		case "--evaluation":
			options.evaluation = args[index]
		case "--activity":
			options.activity = args[index]
		case "--inputs":
			options.inputs = args[index]
		case "--expected-outputs":
			options.expectedOutputs = args[index]
		case "--out":
			options.outputDir = args[index]
		default:
			return verifyRevisionContractOptions{}, fmt.Errorf("%s", verifyRevisionContractUsage)
		}
	}
	if len(positionals) != 2 || options.revision == "" || options.evaluation == "" || options.activity == "" || options.inputs == "" || options.outputDir == "" {
		return verifyRevisionContractOptions{}, fmt.Errorf("%s", verifyRevisionContractUsage)
	}
	options.baseline, options.candidate = positionals[0], positionals[1]
	return options, nil
}
