package main

import (
	"fmt"

	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func compareSavedReports(cfg config) error {
	if cfg.compare == "" || cfg.check == "" || cfg.head == "" {
		return fmt.Errorf("compare, check, and expected-head are required for saved report comparison")
	}
	if cfg.root != "" || cfg.executable != "" || cfg.concept != "" || cfg.corpus != "" || cfg.output != "" {
		return fmt.Errorf("saved comparison accepts no execution or output configuration")
	}
	first, err := readComparedReport(cfg.compare)
	if err != nil {
		return fmt.Errorf("read first report: %w", err)
	}
	replay, err := readComparedReport(cfg.check)
	if err != nil {
		return fmt.Errorf("read replay report: %w", err)
	}
	return metaff.CompareReplayReports(first, replay, cfg.head)
}
