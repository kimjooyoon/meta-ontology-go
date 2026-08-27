package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalcompat"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.input == "" || cfg.expectedRepository == "" || cfg.expectedSHA == "" ||
		cfg.expectedPredecessorSHA == "" ||
		cfg.legacy == "" || cfg.receipt == "" {
		return fmt.Errorf("root, input, expected-repository, expected-sha, expected-predecessor-sha, legacy, and receipt are required")
	}
	if err := requireExternal(cfg.root, cfg.legacy, cfg.receipt); err != nil {
		return err
	}
	raw, err := os.ReadFile(cfg.input)
	if err != nil {
		return err
	}
	bundle, err := proposalcompat.Build(raw, cfg.expectedRepository, cfg.expectedSHA, cfg.expectedPredecessorSHA)
	if err != nil {
		return err
	}
	legacy := proposalcompat.EncodeLegacy(bundle.Legacy)
	receipt := proposalcompat.EncodeReport(bundle.Receipt)
	if cfg.check {
		err = compareFile(cfg.legacy, legacy)
		if err == nil {
			err = compareFile(cfg.receipt, receipt)
		}
	} else {
		err = writeExclusive(cfg.legacy, legacy)
		if err == nil {
			err = writeExclusive(cfg.receipt, receipt)
		}
	}
	if err != nil {
		return err
	}
	fmt.Printf("proposal-compat: decision=%s fields=%d losses=%d writes=%d digest=%s\n",
		bundle.Receipt.Decision, bundle.Receipt.Summary.ProjectedFields,
		bundle.Receipt.Summary.FieldLosses, bundle.Receipt.RepositoryWrites,
		bundle.Receipt.ReportDigest)
	return nil
}
