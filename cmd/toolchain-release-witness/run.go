package main

import (
	"fmt"
	"io"
	"os"

	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.root == "" || cfg.expectedHead == "" || cfg.concept == "" ||
		cfg.corpus == "" || cfg.receipts == "" || cfg.bundle == "" {
		return fmt.Errorf("root, head, concept, corpus, receipts, and bundle are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	if err := requireExternal(cfg.root, cfg.receipts, cfg.bundle, target); err != nil {
		return err
	}
	raw, err := os.ReadFile(cfg.corpus)
	if err != nil {
		return err
	}
	corpus, corpusDigest, err := release.DecodeCorpus(raw)
	if err != nil {
		return err
	}
	conceptDigest, conceptBound, err := readConcept(cfg.concept)
	if err != nil {
		return err
	}
	evidence, err := release.LoadEvidence(cfg.receipts)
	if err != nil {
		return err
	}
	report, err := release.Evaluate(corpus, corpusDigest, evidence,
		cfg.expectedHead, conceptDigest, conceptBound)
	if err != nil {
		return err
	}
	if cfg.check != "" {
		return checkReport(cfg, report, evidence, stdout)
	}
	return produceReport(cfg, report, evidence, stdout)
}
