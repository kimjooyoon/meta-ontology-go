package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxInputBytes = 8 << 20

func readBounded(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxInputBytes {
		return nil, fmt.Errorf("input is not a bounded regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, maxInputBytes+1))
	if err != nil {
		return nil, err
	}
	if len(payload) > maxInputBytes {
		return nil, fmt.Errorf("input exceeds %d bytes", maxInputBytes)
	}
	return payload, nil
}

func writeArtifacts(directory string, source, program, verification []byte) error {
	if info, err := os.Lstat(directory); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output directory is a symbolic link")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return err
	}
	artifacts := []struct {
		name    string
		payload []byte
	}{
		{name: "program.gooo", payload: source},
		{name: "program.json", payload: program},
		{name: "verification.json", payload: verification},
	}
	for _, artifact := range artifacts {
		if err := writeAtomic(filepath.Join(directory, artifact.name), artifact.payload); err != nil {
			return err
		}
	}
	return nil
}

func writeAtomic(path string, payload []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".metric-meta-program-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o640); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
