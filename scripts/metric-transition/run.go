package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metrictransition"
)

func run(args []string) error {
	cfg, err := parseConfig(args)
	if err != nil {
		return err
	}
	options := metrictransition.Options{MetricsPath: cfg.metrics, EffectPath: cfg.effect, ReceiptsPath: cfg.receipts, ProvenancePath: cfg.provenance, PatchPath: cfg.patch, ExpectedSHA: cfg.expected, CIRunID: cfg.ciRun}
	if cfg.verifyState != "" {
		return metrictransition.VerifyFiles(options, cfg.verifyState, cfg.verifyLedger)
	}
	result, err := metrictransition.Build(options)
	if err != nil {
		return err
	}
	if err := metrictransition.WriteResult(result, cfg.outputState, cfg.outputLedger); err != nil {
		return err
	}
	fmt.Printf("metric-transition: decision=%s indicators=%d state=%s\n", result.Ledger.Decision, len(result.Ledger.Indicators), result.State.Digest)
	return nil
}
