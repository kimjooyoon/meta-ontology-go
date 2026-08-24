package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func readTransaction(path string) (transaction, error) {
	file, err := os.Open(path)
	if err != nil { return transaction{}, err }
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var value transaction
	if err := decoder.Decode(&value); err != nil { return transaction{}, err }
	if err := decoder.Decode(&struct{}{}); err != io.EOF { return transaction{}, fmt.Errorf("transaction must contain exactly one JSON value") }
	return value, nil
}

func writeReceipt(path string, receipt rawReceipt) error {
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil { return err }
	raw = append(raw, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil { return err }
	if _, err = file.Write(raw); err != nil { _ = file.Close(); return err }
	return file.Close()
}

func requireExternalOutput(root, output string) error {
	rootPath, err := filepath.EvalSymlinks(root)
	if err != nil { return err }
	rootPath, err = filepath.Abs(rootPath)
	if err != nil { return err }
	parent, err := filepath.EvalSymlinks(filepath.Dir(output))
	if err != nil { return err }
	outputPath, err := filepath.Abs(filepath.Join(parent, filepath.Base(output)))
	if err != nil { return err }
	relative, err := filepath.Rel(rootPath, outputPath)
	if err != nil { return err }
	if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))) { return fmt.Errorf("receipt output must be outside the repository") }
	return nil
}
