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
	members, err := transportMembers(archive)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	var total int64
	for _, name := range transportNames() {
		member := members[name]
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
		actual, err := readBounded(filepath.Join(cfg.Input, name), 32<<20)
		if err != nil || !bytes.Equal(actual, content) {
			return nil, fmt.Errorf("unpacked producer member differs: %s", name)
		}
		files[name] = content
	}
	return files, nil
}

func transportMembers(archive *zip.Reader) (map[string]*zip.File, error) {
	if len(archive.File) != len(transportNames()) {
		return nil, fmt.Errorf("producer archive member population differs")
	}
	allowed := map[string]bool{}
	for _, name := range transportNames() {
		allowed[name] = true
	}
	members := map[string]*zip.File{}
	for _, member := range archive.File {
		if !allowed[member.Name] || members[member.Name] != nil || member.UncompressedSize64 > 32<<20 {
			return nil, fmt.Errorf("producer archive has an unbound member")
		}
		members[member.Name] = member
	}
	return members, nil
}
