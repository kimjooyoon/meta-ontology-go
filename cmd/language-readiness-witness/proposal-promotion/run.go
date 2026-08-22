package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

func run(cfg config) error {
	if cfg.root == "" || cfg.repository == "" || cfg.currentHead == "" ||
		cfg.predecessorSHA == "" || cfg.token == "" {
		return fmt.Errorf("root, repository, current-head, predecessor-sha, and GITHUB_TOKEN are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	if err := requireExternal(cfg.root, target); err != nil {
		return err
	}
	receipt, err := build(cfg)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if cfg.output != "" {
		if err := writeExclusive(cfg.output, data); err != nil {
			return err
		}
	} else {
		actual, err := os.ReadFile(cfg.check)
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, data) {
			return fmt.Errorf("FAIL_CLOSED: proposal promotion replay mismatch")
		}
	}
	fmt.Printf("proposal-promotion: decision=%s coordinates=%d/%d bps=%d writes=%d digest=%s\n",
		receipt.Decision, receipt.Summary.Satisfied, receipt.Summary.Total,
		receipt.Summary.ReadinessBPS, receipt.RepositoryWrites, receipt.ReportDigest)
	return nil
}

func build(cfg config) (proposalpromotion.Receipt, error) {
	collection, err := proposalpredecessor.Collect(
		context.Background(), http.DefaultClient, cfg.apiURL, cfg.token,
		cfg.repository, cfg.predecessorSHA,
	)
	if err != nil {
		return proposalpromotion.Receipt{}, err
	}
	selection, contract, err := proposalpredecessor.Select(
		cfg.repository, cfg.currentHead, cfg.predecessorSHA, collection,
	)
	if err != nil {
		return proposalpromotion.Receipt{}, err
	}
	return proposalpromotion.Build(cfg.currentHead, cfg.predecessorSHA, selection, contract)
}
