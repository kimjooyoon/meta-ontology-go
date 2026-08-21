package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var input, output, expectedSHA, verify string
	flag.StringVar(&input, "input", "", "source metrics report")
	flag.StringVar(&output, "output", "", "witness ledger output")
	flag.StringVar(&expectedSHA, "expected-sha", "", "exact source commit")
	flag.StringVar(&verify, "verify", "", "verify an existing ledger")
	flag.Parse()
	if verify != "" {
		ledger, err := readLedger(verify)
		if err != nil {
			log.Fatal(err)
		}
		if err := validateLedger(ledger); err != nil {
			log.Fatal(err)
		}
		return
	}
	if input == "" || output == "" || expectedSHA == "" {
		log.Fatal("-input, -output, and -expected-sha are required")
	}
	report, err := loadSource(input)
	if err != nil {
		log.Fatal(err)
	}
	ledger, err := buildLedger(report, expectedSHA)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeJSON(output, ledger); err != nil {
		log.Fatal(err)
	}
}

func readLedger(path string) (witnessLedger, error) {
	var ledger witnessLedger
	data, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(data, &ledger)
	}
	return ledger, err
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write ledger: %w", err)
	}
	return nil
}
