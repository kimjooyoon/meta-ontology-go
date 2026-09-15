package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
)

func produce(path string, receipt guardedcapability.Receipt, stdout io.Writer) error {
	raw, err := marshal(receipt)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(raw); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := guardedcapability.ValidateForHead(receipt, receipt.Source.CurrentHeadSHA); err != nil {
		return err
	}
	return printReceipt(stdout, receipt)
}

func printReceipt(stdout io.Writer, receipt guardedcapability.Receipt) error {
	if receipt.Decision != guardedcapability.DecisionPass {
		return fmt.Errorf("guarded capability decision = %s reason = %s", receipt.Decision, receipt.Reason)
	}
	_, err := fmt.Fprintf(stdout,
		"guarded-capability: decision=%s satisfied=%d total=%d bps=%d writes=%d digest=%s\n",
		receipt.Decision, receipt.Summary.Satisfied, receipt.Summary.Total,
		receipt.Summary.ReadinessBPS, receipt.Summary.RepositoryWrites, receipt.ReportDigest)
	return err
}

func marshal(receipt guardedcapability.Receipt) ([]byte, error) {
	raw, err := json.MarshalIndent(receipt, "", "  ")
	return append(raw, '\n'), err
}
