package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "COMMAND", Reason: "MISSING_COMMAND"})
	}
	command := os.Args[1]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	root := flags.String("root", ".", "repository root")
	output := flags.String("output", "", "generated output directory")
	evidencePath := flags.String("evidence", "", "proof evidence output path")
	if err := flags.Parse(os.Args[2:]); err != nil {
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "COMMAND", Reason: "INVALID_COMMAND_FLAGS"})
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "ROOT", Reason: "INVALID_REPOSITORY_ROOT"})
	}
	outputDir := *output
	if outputDir == "" {
		outputDir = filepath.Join(absoluteRoot, defaultOutput)
	} else if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(absoluteRoot, outputDir)
	}
	switch command {
	case "generate":
		runGenerate(absoluteRoot, outputDir)
	case "check":
		runCheck(absoluteRoot, outputDir)
	case "prove":
		runProve(absoluteRoot, outputDir, *evidencePath)
	default:
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "COMMAND", Reason: "UNKNOWN_COMMAND"})
	}
}

func runGenerate(root, outputDir string) {
	loaded, err := loadManifests(root)
	if err != nil {
		failError(err, "FOUNDATION", "LOAD_MANIFESTS", "MANIFEST_LOAD_FAILED")
	}
	outputs, _, err := renderOutputs(root, outputDir, loaded)
	if err != nil {
		failError(err, "FOUNDATION", "BUILD_PROJECTION", "PROJECTION_BUILD_FAILED")
	}
	if err := writeOutputs(outputDir, outputs); err != nil {
		failError(err, "REGRESSION", "GENERATED_OUTPUT", "GENERATED_OUTPUT_WRITE_FAILED")
	}
	printJSON(map[string]any{"decision": "PASS", "stage": "COHERENCE", "step": "GENERATE", "reason": "GENERATED_PROJECTION", "output": filepath.ToSlash(outputDir), "files": len(outputs)})
}

func runCheck(root, outputDir string) {
	loaded, err := loadManifests(root)
	if err != nil {
		failError(err, "FOUNDATION", "LOAD_MANIFESTS", "MANIFEST_LOAD_FAILED")
	}
	expectedIDs, diagnostic := expectedManifestIDs(outputDir)
	if diagnostic != nil {
		fail(diagnostic)
	}
	if diagnostic := validateManifestInputs(root, loaded, expectedIDs); diagnostic != nil {
		fail(diagnostic)
	}
	expected, _, err := renderOutputs(root, outputDir, loaded)
	if err != nil {
		failError(err, "FOUNDATION", "BUILD_PROJECTION", "PROJECTION_BUILD_FAILED")
	}
	observed, err := readGenerated(outputDir)
	if err != nil {
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "GENERATED_OUTPUT", Reason: "MISSING_GENERATED_PROJECTION"})
	}
	if diagnostic := checkRendered(expected, observed); diagnostic != nil {
		fail(diagnostic)
	}
	printJSON(map[string]any{"decision": "PASS", "stage": "REGRESSION", "step": "GENERATED_OUTPUT", "reason": "GENERATED_PROJECTION_FRESH", "output": filepath.ToSlash(outputDir)})
}

func runProve(root, outputDir, evidencePath string) {
	evidence, err := runProof(root, outputDir)
	if err != nil {
		failError(err, "FOUNDATION", "PROVE", "PROOF_EXECUTION_FAILED")
	}
	if evidencePath != "" {
		data, marshalErr := renderJSON(evidence)
		if marshalErr != nil {
			failError(marshalErr, "REGRESSION", "EVIDENCE", "EVIDENCE_RENDER_FAILED")
		}
		if writeErr := os.WriteFile(evidencePath, data, 0o644); writeErr != nil {
			failError(writeErr, "REGRESSION", "EVIDENCE", "EVIDENCE_WRITE_FAILED")
		}
	}
	printJSON(evidence)
	if evidence.Decision != "PASS" {
		os.Exit(1)
	}
}

func failError(err error, stage, step, reason string) {
	if diagnostic, ok := err.(diagnosticFailure); ok {
		fail(&diagnostic.Diagnostic)
	}
	fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: stage, Step: step, Reason: reason})
}

func printJSON(value any) {
	data, err := renderJSON(value)
	if err != nil {
		fail(&Diagnostic{Decision: "FAIL_CLOSED", Stage: "REGRESSION", Step: "OUTPUT", Reason: "OUTPUT_RENDER_FAILED"})
	}
	_, _ = os.Stdout.Write(data)
}

type diagnosticFailure struct {
	Diagnostic
}

func (failure diagnosticFailure) Error() string {
	return failure.Decision + "/" + failure.Stage + "/" + failure.Reason
}

func fail(diagnostic *Diagnostic) {
	data, _ := renderJSON(diagnostic)
	_, _ = os.Stderr.Write(data)
	os.Exit(1)
}
