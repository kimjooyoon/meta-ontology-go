package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

type config struct {
	root, metrics, plan, execution, receipts, provenance, expected string
	output, generatedReceipts, executedProvenance, patch, verify   string
}

func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	if cfg.verify != "" {
		return transformationeffect.VerifyFiles(cfg.verify, cfg.generatedReceipts, cfg.executedProvenance, cfg.patch)
	}
	invocation, err := invocationID(cfg.output)
	if err != nil {
		return err
	}
	outputDir := filepath.Dir(cfg.output)
	result, err := transformationeffect.Build(transformationeffect.Options{
		Root: cfg.root, MetricsPath: cfg.metrics, PlanPath: cfg.plan, ExecutionPath: cfg.execution,
		ReceiptsPath: cfg.receipts, ProvenancePath: cfg.provenance, ExpectedSHA: cfg.expected,
		OutputPath: cfg.output, ProgressPath: filepath.Join(outputDir, "operation-progress.jsonl"),
		InvocationID: invocation,
	})
	if err != nil {
		if diagnosticErr := transformationeffect.WriteReplayDiagnostic(cfg.output, err); diagnosticErr != nil {
			return fmt.Errorf("%w; write replay diagnostic: %v", err, diagnosticErr)
		}
		return err
	}
	if err := transformationeffect.WriteResult(result, cfg.output, cfg.generatedReceipts, cfg.executedProvenance, cfg.patch); err != nil {
		return err
	}
	fmt.Printf("transformation-effect: decision=%s effects=%d status=%s\n",
		result.Ledger.Decision, len(result.Ledger.Effects), result.Ledger.Status)
	return nil
}

func invocationID(output string) (string, error) {
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("generate diagnostic invocation identity: %w", err)
	}
	runID := os.Getenv("GITHUB_RUN_ID")
	attempt := os.Getenv("GITHUB_RUN_ATTEMPT")
	job := os.Getenv("GITHUB_JOB")
	if runID == "" {
		runID = "local"
	}
	if attempt == "" {
		attempt = "1"
	}
	if job == "" {
		job = "transformation-effect"
	}
	return fmt.Sprintf("%s/%s/%s/%s:%s", runID, attempt, job, hex.EncodeToString(token), output), nil
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("transformation-effect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.root, "root", ".", "materialized source workspace")
	flags.StringVar(&cfg.metrics, "metrics", "", "source metrics JSON")
	flags.StringVar(&cfg.plan, "plan", "", "generation plan JSON")
	flags.StringVar(&cfg.execution, "execution", "", "execution manifest JSON")
	flags.StringVar(&cfg.receipts, "receipts", "", "input receipt report JSON")
	flags.StringVar(&cfg.provenance, "provenance", "", "input provenance JSON")
	flags.StringVar(&cfg.expected, "expected-sha", "", "exact source SHA")
	flags.StringVar(&cfg.output, "output", "", "effect ledger output")
	flags.StringVar(&cfg.generatedReceipts, "generated-receipts", "", "executed receipt output")
	flags.StringVar(&cfg.executedProvenance, "executed-provenance", "", "executed provenance output")
	flags.StringVar(&cfg.patch, "patch", "", "content patch output")
	flags.StringVar(&cfg.verify, "verify", "", "verify an existing effect ledger")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return cfg, fmt.Errorf("invalid arguments")
	}
	outputs := cfg.generatedReceipts != "" && cfg.executedProvenance != "" && cfg.patch != ""
	inputs := cfg.metrics != "" && cfg.plan != "" && cfg.execution != "" && cfg.receipts != "" && cfg.provenance != ""
	if !outputs || cfg.verify == "" && (!inputs || cfg.expected == "" || cfg.output == "") {
		return cfg, fmt.Errorf("exact input and output paths are required")
	}
	return cfg, nil
}
