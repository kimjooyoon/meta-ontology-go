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

const reviseFromHandoffUsage = "usage: gooo revise-from-handoff <source.gooo> --handoff <repair-handoff.json> --source-digest <sha256:...> --activity <name> --expected <program> --replace <program> --out <directory>"

type reviseFromHandoffOptions struct {
	filename, handoff, sourceDigest, activity, expected, replacement, outputDir string
}

func runReviseFromHandoff(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, err := parseReviseFromHandoffArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, reviseFromHandoffUsage)
		return exitUsage
	}
	source, err := readSource(reader, options.filename)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	handoffData, err := reader.ReadFile(options.handoff)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	var handoff valueexecution.RepairHandoff
	if err := json.Unmarshal(handoffData, &handoff); err != nil {
		fmt.Fprintf(stderr, "gooo: revise-from-handoff: decode handoff: %v\n", err)
		return exitFailure
	}
	candidate, revision, err := valueexecution.ProposeSourceRevisionFromHandoff(options.filename, source, handoff, valueexecution.SourceRevisionRequest{
		SourceDigest: options.sourceDigest, Activity: options.activity, ExpectedProgram: options.expected, ReplacementProgram: options.replacement,
	})
	if err != nil {
		fmt.Fprintf(stderr, "gooo: revise-from-handoff: %v\n", err)
		return exitFailure
	}
	payload, err := json.MarshalIndent(revision, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "gooo: revise-from-handoff: encode revision: %v\n", err)
		return exitFailure
	}
	payload = append(payload, '\n')
	if err := writeHandoffSourceRevisionArtifacts(options.outputDir, candidate, payload, handoffData); err != nil {
		fmt.Fprintf(stderr, "gooo: revise-from-handoff: output: %v\n", err)
		return exitFailure
	}
	fmt.Fprintf(stdout, "source revision from handoff: %s\n", filepath.Join(options.outputDir, "candidate.gooo"))
	return exitOK
}

func parseReviseFromHandoffArguments(args []string) (reviseFromHandoffOptions, error) {
	var options reviseFromHandoffOptions
	seen := make(map[string]bool)
	for index := 0; index < len(args); index++ {
		if !strings.HasPrefix(args[index], "--") {
			if options.filename != "" {
				return reviseFromHandoffOptions{}, fmt.Errorf("%s", reviseFromHandoffUsage)
			}
			options.filename = args[index]
			continue
		}
		if index+1 >= len(args) || seen[args[index]] {
			return reviseFromHandoffOptions{}, fmt.Errorf("%s", reviseFromHandoffUsage)
		}
		seen[args[index]] = true
		index++
		switch args[index-1] {
		case "--handoff":
			options.handoff = args[index]
		case "--source-digest":
			options.sourceDigest = args[index]
		case "--activity":
			options.activity = args[index]
		case "--expected":
			options.expected = args[index]
		case "--replace":
			options.replacement = args[index]
		case "--out":
			options.outputDir = args[index]
		default:
			return reviseFromHandoffOptions{}, fmt.Errorf("%s", reviseFromHandoffUsage)
		}
	}
	if options.filename == "" || options.handoff == "" || options.sourceDigest == "" || options.activity == "" || options.expected == "" || options.replacement == "" || options.outputDir == "" {
		return reviseFromHandoffOptions{}, fmt.Errorf("%s", reviseFromHandoffUsage)
	}
	return options, nil
}

func writeHandoffSourceRevisionArtifacts(outputDir string, candidate, revision, handoff []byte) error {
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
	if err := os.WriteFile(filepath.Join(outputDir, "revision.json"), revision, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "repair-handoff.json"), handoff, 0o644)
}
