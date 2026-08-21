package main

import (
	"flag"
	"fmt"
	"io"
)

type config struct {
	metrics, effect, receipts, provenance, patch, expected, ciRun string
	outputState, outputLedger, verifyState, verifyLedger          string
}

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("metric-transition", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.metrics, "metrics", "", "source metrics JSON")
	flags.StringVar(&cfg.effect, "effect", "", "transformation effect ledger")
	flags.StringVar(&cfg.receipts, "receipts", "", "executed receipts")
	flags.StringVar(&cfg.provenance, "provenance", "", "executed provenance")
	flags.StringVar(&cfg.patch, "patch", "", "content patch")
	flags.StringVar(&cfg.expected, "expected-sha", "", "exact source SHA")
	flags.StringVar(&cfg.ciRun, "ci-run-id", "", "source CI run")
	flags.StringVar(&cfg.outputState, "output-state", "", "canonical state output")
	flags.StringVar(&cfg.outputLedger, "output-ledger", "", "transition ledger output")
	flags.StringVar(&cfg.verifyState, "verify-state", "", "state to replay")
	flags.StringVar(&cfg.verifyLedger, "verify-ledger", "", "ledger to replay")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return cfg, fmt.Errorf("invalid arguments")
	}
	inputs := cfg.metrics != "" && cfg.effect != "" && cfg.receipts != "" && cfg.provenance != "" && cfg.patch != "" && cfg.expected != "" && cfg.ciRun != ""
	build := cfg.outputState != "" && cfg.outputLedger != "" && cfg.verifyState == "" && cfg.verifyLedger == ""
	verify := cfg.verifyState != "" && cfg.verifyLedger != "" && cfg.outputState == "" && cfg.outputLedger == ""
	if !inputs || build == verify {
		return cfg, fmt.Errorf("exact inputs and one output mode are required")
	}
	return cfg, nil
}
