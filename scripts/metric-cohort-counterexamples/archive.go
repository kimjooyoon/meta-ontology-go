package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
)

func transportNames() []string {
	return []string{"concept-operation-binding.json", "strategy-plan.json", "strategy-verification.json",
		"source-metrics.json", "intervention-ledger.json", "intervention-verification.json",
		"program.json", "program.gooo", "verification.json", "producer-attempt.json"}
}

func transportPayloads(cfg config) (map[string][]byte, error) {
	raw, err := boundArchive(cfg)
	if err != nil {
		return nil, err
	}
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	if len(archive.File) != len(transportNames()) {
		return nil, fmt.Errorf("producer archive member population differs")
	}
	allowed := map[string]bool{}
	for _, name := range transportNames() {
		allowed[name] = true
	}
	files := map[string][]byte{}
	var total int64
	for _, member := range archive.File {
		_, duplicate := files[member.Name]
		if !allowed[member.Name] || duplicate || member.UncompressedSize64 > 32<<20 {
			return nil, fmt.Errorf("producer archive has an unbound member")
		}
		reader, err := member.Open()
		if err != nil {
			return nil, err
		}
		content, readErr := io.ReadAll(io.LimitReader(reader, (32<<20)+1))
		closeErr := reader.Close()
		total += int64(len(content))
		if readErr != nil || closeErr != nil || len(content) > 32<<20 || total > 64<<20 {
			return nil, fmt.Errorf("producer archive cannot be read within its member budget")
		}
		actual, err := readBounded(filepath.Join(cfg.Input, member.Name), 32<<20)
		if err != nil || !bytes.Equal(actual, content) {
			return nil, fmt.Errorf("unpacked producer member differs: %s", member.Name)
		}
		files[member.Name] = content
	}
	return files, nil
}
