package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance"
)

func readTransaction(path string) (languageassurance.Transaction, error) {
	file, err := os.Open(path)
	if err != nil {
		return languageassurance.Transaction{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var transaction languageassurance.Transaction
	if err := decoder.Decode(&transaction); err != nil {
		return languageassurance.Transaction{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return languageassurance.Transaction{}, fmt.Errorf("transaction must contain exactly one JSON value")
	}
	return transaction, nil
}

func writeReport(path string, report languageassurance.Report) error {
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(raw); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func requireExternalOutput(root, output string) error {
	rootPath, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(output))
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs(filepath.Join(parent, filepath.Base(output)))
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil {
		return err
	}
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))) {
		return fmt.Errorf("report output must be outside the repository")
	}
	return nil
}
