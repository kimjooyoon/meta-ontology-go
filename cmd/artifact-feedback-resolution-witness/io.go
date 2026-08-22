package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactfeedback"
)

func readInput(path string) (artifactfeedback.ResolutionInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return artifactfeedback.ResolutionInput{}, err
	}
	var input artifactfeedback.ResolutionInput
	if err := json.Unmarshal(data, &input); err != nil {
		return artifactfeedback.ResolutionInput{}, err
	}
	return input, nil
}

func newReceipt(report artifactfeedback.ResolutionReport, replayDigest, expected string) (receipt, error) {
	output := receipt{
		Schema: receiptSchema, Report: report, ReplayReportDigest: replayDigest,
		ExpectedDigest: expected, ReplayVerified: report.ReportDigest == replayDigest,
		RepositoryWrites: report.RepositoryWrites,
	}
	digest, err := digestJSON(output)
	if err != nil {
		return receipt{}, err
	}
	output.ReceiptDigest = digest
	return output, nil
}

func marshalReceipt(output receipt) ([]byte, error) {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create resolution receipt: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
