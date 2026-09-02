package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func runTransport(input options) error {
	raw, err := os.ReadFile(input.receipt)
	if err != nil {
		return fmt.Errorf("read lifecycle receipt: %w", err)
	}
	var receipt selfimprovementtransport.LifecycleReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return fmt.Errorf("decode lifecycle receipt: %w", err)
	}
	archiveRaw, downloadExit := readLookup(input.archive, input.downloadExit)
	receipt = selfimprovementtransport.CompleteArtifactLifecycle(receipt, archiveRaw, downloadExit)
	if err := writeReceipt(input.receipt, receipt); err != nil {
		return err
	}
	printReceipt(receipt)
	return nil
}

func readLookup(path string, exitCode int) ([]byte, int) {
	if exitCode != 0 {
		return nil, exitCode
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 1
	}
	return raw, 0
}
