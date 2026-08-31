package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	readinessartifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact"
)

func produce(cfg config, stdout io.Writer) error {
	receipt, err := build(cfg)
	if err != nil {
		return err
	}
	data, err := marshalReceipt(receipt)
	if err != nil {
		return err
	}
	if err := writeExclusive(cfg.output, data); err != nil {
		return err
	}
	printSummary(stdout, receipt)
	return nil
}

func consume(cfg config, stdout io.Writer) error {
	data, err := os.ReadFile(cfg.check)
	if err != nil {
		return err
	}
	actual := readinessartifact.Receipt{}
	if err := json.Unmarshal(data, &actual); err != nil {
		return err
	}
	if err := readinessartifact.Validate(actual); err != nil {
		return err
	}
	expected, err := build(cfg)
	if err != nil {
		return err
	}
	actualData, _ := marshalReceipt(actual)
	expectedData, _ := marshalReceipt(expected)
	if !bytes.Equal(actualData, expectedData) {
		return fmt.Errorf("FAIL_CLOSED: readiness artifact does not match exact input")
	}
	printSummary(stdout, actual)
	return nil
}
