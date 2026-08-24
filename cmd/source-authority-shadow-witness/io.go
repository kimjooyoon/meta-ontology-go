package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityshadow"
)

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
		return fmt.Errorf("shadow receipt output must be outside the repository")
	}
	return nil
}

func writeReceipt(path string, receipt sourceauthorityshadow.Receipt) error {
	raw, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(append(raw, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
