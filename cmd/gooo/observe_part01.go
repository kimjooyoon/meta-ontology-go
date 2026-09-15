package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const observeUsage = "usage: gooo observe <observation.gooo> --input <file.gooo> --out <directory>"

type observeOptions struct {
	contractFilename string
	inputFilename    string
	outputDir        string
}

type observationInputs struct {
	contractSource []byte
	inputSource    []byte
	contract       generation.SemanticObservationContract
	file           *syntax.File
}

func runObserve(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseObserveArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	inputs, diagnostics, err := loadObservationInputs(options, reader, parser)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if diagnostics.HasErrors() {
		_ = reportDiagnostics(diagnostics, stderr)
		return exitFailure
	}
	if !reportDiagnostics(diagnostics, stderr) {
		return exitFailure
	}

	recorder := generation.NewSemanticObservationRecorder()
	observation, err := observeCompilerPair(options, inputs, recorder)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: observation report: %v\n", err)
		return exitFailure
	}
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: observation report: %v\n", err)
		return exitFailure
	}
	data = append(data, '\n')
	if err := writeObservationArtifact(options.outputDir, data); err != nil {
		fmt.Fprintf(stderr, "gooo: observation output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "observed: %s\n", filepath.Join(options.outputDir, "compiler-observation.json"))
	return exitOK
}

func loadObservationInputs(options observeOptions, reader SourceReader, parser SourceParser) (observationInputs, syntax.Diagnostics, error) {
	contractSource, err := reader.ReadFile(options.contractFilename)
	if err != nil {
		return observationInputs{}, nil, fmt.Errorf("gooo: %s: read error: %v", options.contractFilename, err)
	}
	contract, err := generation.ParseSemanticObservationContract(contractSource)
	if err != nil {
		return observationInputs{}, nil, fmt.Errorf("gooo: %s: observation contract: %v", options.contractFilename, err)
	}
	inputSource, err := reader.ReadFile(options.inputFilename)
	if err != nil {
		return observationInputs{}, nil, fmt.Errorf("gooo: %s: read error: %v", options.inputFilename, err)
	}
	file, diagnostics, err := parseWithDeadline(parser, options.inputFilename, string(inputSource), commandDeadline)
	if err != nil {
		return observationInputs{}, nil, fmt.Errorf("gooo: %s: parse error: %v", options.inputFilename, err)
	}
	return observationInputs{contractSource: contractSource, inputSource: inputSource, contract: contract, file: file}, diagnostics, nil
}

func observeCompilerPair(options observeOptions, inputs observationInputs, recorder *generation.SemanticObservationRecorder) (generation.SemanticObservation, error) {
	var beforeMem, afterMem runtime.MemStats
	runtime.ReadMemStats(&beforeMem)
	started := time.Now()
	first, err := generateWithDeadlineObserved(inputs.file, nil, commandDeadline, recorder)
	if err != nil {
		return generation.SemanticObservation{}, fmt.Errorf("gooo: %s: observed generation failed: %v", options.inputFilename, err)
	}
	second, err := generateWithDeadlineObserved(inputs.file, nil, commandDeadline, recorder)
	if err != nil {
		return generation.SemanticObservation{}, fmt.Errorf("gooo: %s: observed replay failed: %v", options.inputFilename, err)
	}
	runtime.ReadMemStats(&afterMem)
	behaviorEqual := bytes.Equal(first.result.Source, second.result.Source) && first.ir.StableHash() == second.ir.StableHash()
	pair := generation.SemanticObservationPairEvidence{EvidenceAvailable: true, ChangeAdopted: false, BehaviorEqual: behaviorEqual, DeterminismEqual: behaviorEqual, BeforeOperationCount: 2, AfterOperationCount: 2}
	observation, err := recorder.BuildSemanticObservation(inputs.contract, cache.HashBytes(inputs.contractSource).String(), cache.HashBytes(inputs.inputSource).String(), pair)
	if err != nil {
		return generation.SemanticObservation{}, err
	}
	observation.Metrics.InputGoooPhysicalLines = countObservationPhysicalLines(inputs.inputSource)
	observation.Metrics.OutputArtifactFiles = 1
	observation.Metrics.AllocationCount = int64(afterMem.Mallocs - beforeMem.Mallocs)
	observation.Metrics.AllocationBytes = int64(afterMem.TotalAlloc - beforeMem.TotalAlloc)
	observation.Metrics.WallMS = int64((time.Since(started).Nanoseconds() + int64(time.Millisecond) - 1) / int64(time.Millisecond))
	observation.Metrics.RepositoryWrites = 0
	observation.Metrics.LocalTestExecutions = 0
	if !behaviorEqual {
		observation.Decision = "REFUTED"
		observation.Reason = "BEHAVIOR_OR_DETERMINISM_MISMATCH"
		observation.PairEvidence.Contradiction = "identical compiler inputs produced different generated behavior"
	}
	return observation, nil
}

func parseObserveArguments(args []string) (observeOptions, error) {
	if len(args) == 0 {
		return observeOptions{}, fmt.Errorf("%s", observeUsage)
	}
	options := observeOptions{contractFilename: args[0]}
	for index := 1; index < len(args); index++ {
		if index+1 >= len(args) || args[index+1] == "" {
			return observeOptions{}, fmt.Errorf("%s", observeUsage)
		}
		switch args[index] {
		case "--input":
			if options.inputFilename != "" {
				return observeOptions{}, fmt.Errorf("%s", observeUsage)
			}
			options.inputFilename = args[index+1]
		case "--out":
			if options.outputDir != "" {
				return observeOptions{}, fmt.Errorf("%s", observeUsage)
			}
			options.outputDir = args[index+1]
		default:
			return observeOptions{}, fmt.Errorf("%s", observeUsage)
		}
		index++
	}
	if options.inputFilename == "" || options.outputDir == "" {
		return observeOptions{}, fmt.Errorf("%s", observeUsage)
	}
	return options, nil
}

func writeObservationArtifact(outputDir string, data []byte) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create caller-owned output directory: %w", err)
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("inspect caller-owned output directory: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output directory must be empty")
	}
	if err := os.WriteFile(filepath.Join(outputDir, "compiler-observation.json"), data, 0o644); err != nil {
		return fmt.Errorf("write compiler observation: %w", err)
	}
	return nil
}

func countObservationPhysicalLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	count := 1
	for _, value := range source {
		if value == '\n' {
			count++
		}
	}
	if source[len(source)-1] == '\n' {
		count--
	}
	return count
}
