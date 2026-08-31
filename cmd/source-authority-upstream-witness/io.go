package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityupstream"
)

type artifact struct {
	name string
	data []byte
}

func writeArtifacts(outputDir string, suite sourceauthorityupstream.Suite) error {
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	artifacts := make([]artifact, 0, len(suite.Cases)+1)
	for _, result := range suite.Cases {
		encoded, err := encodeJSON(result.Receipt)
		if err != nil {
			return err
		}
		artifacts = append(artifacts, artifact{name: result.ID + ".json", data: encoded})
	}
	encoded, err := encodeJSON(suite)
	if err != nil {
		return err
	}
	artifacts = append(artifacts, artifact{name: "suite.json", data: encoded})
	manifest := make([]string, 0, len(artifacts))
	for _, item := range artifacts {
		if err := writeExclusive(filepath.Join(outputDir, item.name), item.data); err != nil {
			return err
		}
		digest := sha256.Sum256(item.data)
		manifest = append(manifest, hex.EncodeToString(digest[:])+"  "+item.name)
	}
	return writeExclusive(filepath.Join(outputDir, "manifest.sha256"), []byte(strings.Join(manifest, "\n")+"\n"))
}

func encodeJSON(value any) ([]byte, error) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	return append(encoded, '\n'), err
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}
