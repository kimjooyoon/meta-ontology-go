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

const reviseSourceUsage = "usage: gooo revise-source <source.gooo> --source-digest <sha256:...> --activity <name> --expected <program> --replace <program> --reason <reason> --out <directory>"

type reviseSourceOptions struct {
	filename, sourceDigest, activity, expected, replacement, reason, outputDir string
}

func runReviseSource(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseReviseSourceArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, reviseSourceUsage)
		return exitUsage
	}
	source, err := readSource(reader, options.filename)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	candidate, revision, err := valueexecution.ProposeSourceRevision(options.filename, source, valueexecution.SourceRevisionRequest{
		SourceDigest: options.sourceDigest, Activity: options.activity, ExpectedProgram: options.expected,
		ReplacementProgram: options.replacement, TriggerReason: options.reason,
	})
	if err != nil {
		fmt.Fprintf(stderr, "gooo: revise-source: %v\n", err)
		return exitFailure
	}
	payload, err := json.MarshalIndent(revision, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: revise-source: encode: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeSourceRevisionArtifacts(options.outputDir, candidate, payload); err != nil {
		fmt.Fprintf(stderr, "gooo: revise-source: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "source revision candidate: %s\n", filepath.Join(options.outputDir, "candidate.gooo"))
	return exitOK
}

func parseReviseSourceArguments(args []string) (reviseSourceOptions, error) {
	var options reviseSourceOptions
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			if options.filename != "" {
				return reviseSourceOptions{}, fmt.Errorf("%s", reviseSourceUsage)
			}
			options.filename = args[index]
			continue
		}
		if index+1 >= len(args) || seen[args[index]] {
			return reviseSourceOptions{}, fmt.Errorf("%s", reviseSourceUsage)
		}
		seen[args[index]] = true
		index++
		switch args[index-1] {
		case "--source-digest":
			options.sourceDigest = args[index]
		case "--activity":
			options.activity = args[index]
		case "--expected":
			options.expected = args[index]
		case "--replace":
			options.replacement = args[index]
		case "--reason":
			options.reason = args[index]
		case "--out":
			options.outputDir = args[index]
		default:
			return reviseSourceOptions{}, fmt.Errorf("%s", reviseSourceUsage)
		}
	}
	if options.filename == "" || options.sourceDigest == "" || options.activity == "" || options.expected == "" || options.replacement == "" || options.reason == "" || options.outputDir == "" {
		return reviseSourceOptions{}, fmt.Errorf("%s", reviseSourceUsage)
	}
	return options, nil
}

func writeSourceRevisionArtifacts(outputDir string, candidate, revision []byte) error {
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
	if err := os.WriteFile(filepath.Join(outputDir, "candidate.gooo"), candidate, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "revision.json"), revision, 0o644)
}
