package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundary"
)

func main() {
	sourcePath := flag.String("source", metacircularboundary.ExpectedSourcePath, "Gooo source to observe")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	output := flag.String("output", "", "receipt output path")
	grantPath := flag.String("grant", "", "raw external grant artifact path")
	effectPath := flag.String("effect-evidence", "", "raw workspace effect artifact path")
	replayPath := flag.String("replay-evidence", "", "raw replay evidence artifact path")
	executionDir := flag.String("execution-dir", "", "directory for execution artifacts")
	flag.Parse()

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	grant, err := readOptional(*grantPath)
	if err != nil {
		fatal(err)
	}
	effect, err := readOptional(*effectPath)
	if err != nil {
		fatal(err)
	}
	replay, err := readOptional(*replayPath)
	if err != nil {
		fatal(err)
	}
	input := metacircularboundary.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source, GrantEvidence: grant, EffectEvidence: effect, ReplayEvidence: replay}
	initial := metacircularboundary.Evaluate(input)
	for _, item := range initial.Cases {
		if item.Observation.Authorization != metacircularboundary.AuthorizationGranted || !item.Attempt.RequestExecution {
			continue
		}
		artifact, err := metacircularboundary.ExecuteReadOnlyMetaOperation(initial.Source, item.Grant, item.Definition.ID)
		if err != nil {
			fatal(err)
		}
		if *executionDir == "" {
			fatal(fmt.Errorf("execution artifact directory is required for an authorized execution"))
		}
		if err := metacircularboundary.WriteExecutionArtifact(*executionDir, artifact); err != nil {
			fatal(err)
		}
		input.ExecutionArtifacts = append(input.ExecutionArtifacts, artifact)
	}
	report := metacircularboundary.Evaluate(input)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("meta-circular boundary: %s %d/%d auth=%d exec=%d\n", report.Decision, report.Summary.CasesPassed, report.Summary.CasesTotal, report.Summary.ExplicitAuthorizations, report.Summary.AllowedExecutions)
}

func readOptional(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
