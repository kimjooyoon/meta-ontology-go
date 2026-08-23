package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
)

func run(cfg config, stdout io.Writer) error {
	if cfg.printFoundationArtifactID {
		if cfg.root != "" || cfg.currentHead != "" || cfg.foundationArchive != "" ||
			cfg.output != "" || cfg.check != "" {
			return fmt.Errorf("print-foundation-artifact-id does not accept receipt inputs")
		}
		_, err := fmt.Fprintln(stdout, guardedcapability.FoundationArtifactID)
		return err
	}
	if cfg.root == "" || cfg.currentHead == "" || (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("root, current-head, and exactly one of output or check are required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	if err := requireExternal(cfg.root, target, cfg.foundationArchive); err != nil {
		return err
	}
	expected, err := build(cfg)
	if err != nil {
		return err
	}
	if cfg.output != "" {
		return produce(cfg.output, expected, stdout)
	}
	raw, err := os.ReadFile(cfg.check)
	if err != nil {
		return err
	}
	actual := guardedcapability.Receipt{}
	if err := json.Unmarshal(raw, &actual); err != nil {
		return err
	}
	actualRaw, _ := marshal(actual)
	expectedRaw, _ := marshal(expected)
	if err := guardedcapability.ValidateForHead(actual, cfg.currentHead); err != nil ||
		!bytes.Equal(actualRaw, expectedRaw) {
		return fmt.Errorf("FAIL_CLOSED: guarded capability does not match exact input: %v", err)
	}
	return printReceipt(stdout, actual)
}

func build(cfg config) (guardedcapability.Receipt, error) {
	if cfg.foundationArchive != "" {
		raw, err := os.ReadFile(cfg.foundationArchive)
		if err != nil {
			return guardedcapability.Receipt{}, err
		}
		if err := guardedcapability.VerifyFoundationArchive(raw); err != nil {
			return guardedcapability.Receipt{}, err
		}
	}
	source, err := guardedcapability.Collect(cfg.root, cfg.currentHead)
	if err != nil {
		return guardedcapability.Receipt{}, err
	}
	return guardedcapability.Build(source), nil
}
