package main

import (
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityshadow"
)

func run(cfg config) (bool, error) {
	if cfg.root == "" || cfg.input == "" || cfg.output == "" || cfg.expectedSHA == "" {
		return false, fmt.Errorf("root, input, output, and expected-sha are required")
	}
	if err := requireExternalOutput(cfg.root, cfg.output); err != nil {
		return false, err
	}
	raw, err := os.ReadFile(cfg.input)
	if err != nil {
		return false, err
	}
	report := sourceauthorityshadow.Observe(raw, cfg.expectedSHA)
	replay := sourceauthorityshadow.Observe(raw, cfg.expectedSHA)
	if report.ReceiptDigest == "" || report.ReceiptDigest != replay.ReceiptDigest {
		return false, fmt.Errorf("source authority shadow replay digest mismatch")
	}
	if err := writeReceipt(cfg.output, report); err != nil {
		return false, err
	}
	fmt.Printf("source-authority-shadow: observation=%s resolution=%s enforcement=%s digest=%s\n",
		report.Observation, report.Resolution, report.Enforcement, report.ReceiptDigest)
	failClosed := report.Observation != "SATISFIED" || report.Resolution != "EXACT" ||
		report.Enforcement != "ALLOW" || report.GateEffect != "NO_EFFECT" ||
		report.PromotionCreditBPS != 0 || report.RepositoryWrites != 0
	return cfg.check && failClosed, nil
}
