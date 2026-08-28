package main

import (
	"fmt"
	"go/format"
	"os"
)

func formatStaged(root string, buffers map[string][]byte, created map[string]bool) (map[string]stagedFile, error) {
	staged := make(map[string]stagedFile, len(buffers))
	for logical, data := range buffers {
		name, err := extractionPath(root, logical)
		if err != nil {
			return nil, err
		}
		formatted, err := format.Source(data)
		if err != nil {
			return nil, fmt.Errorf("format extraction %s: %w", logical, err)
		}
		if !created[logical] && extractionLines(formatted) > 75 {
			return nil, fmt.Errorf("extraction target %s remains at %d lines", logical, lines)
		}
		mode := uint32(0o644)
		if !created[logical] {
			info, err := os.Stat(name)
			if err != nil {
				return nil, err
			}
			mode = uint32(info.Mode().Perm())
		}
		staged[logical] = stagedFile{name: name, data: formatted, mode: mode, created: created[logical]}
	}
	return staged, nil
}

func installTransaction(file transactionFile) error {
	if file.created {
		if _, err := os.Lstat(file.name); err == nil || !os.IsNotExist(err) {
			return fmt.Errorf("creation target exists: %s", file.name)
		}
		return os.Rename(file.temp, file.name)
	}
	if err := os.Rename(file.name, file.backup); err != nil {
		return err
	}
	if err := os.Rename(file.temp, file.name); err != nil {
		_ = os.Rename(file.backup, file.name)
		return err
	}
	return nil
}

func removeTransactionBackup(file transactionFile) error {
	if file.created {
		return nil
	}
	return os.Remove(file.backup)
}

func restoreTransaction(file transactionFile) {
	if file.created {
		_ = os.Remove(file.name)
		return
	}
	_ = os.Remove(file.name)
	_ = os.Rename(file.backup, file.name)
}
