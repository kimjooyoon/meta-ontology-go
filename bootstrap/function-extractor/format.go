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
		if extractionLines(formatted) > 75 {
			return nil, fmt.Errorf("extraction target %s remains at %d lines", logical, extractionLines(formatted))
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

func installTransaction(file *transactionFile) (namespaceReplacementReceipt, error) {
	if !sameDirectory(file.name, file.temp) {
		return namespaceReplacementReceipt{}, fmt.Errorf("replacement paths are not same-directory: %s", file.logical)
	}
	if file.created {
		if _, err := os.Lstat(file.name); err == nil || !os.IsNotExist(err) {
			return namespaceReplacementReceipt{}, fmt.Errorf("creation target exists: %s", file.name)
		}
	} else if err := preserveDestination(file); err != nil {
		return namespaceReplacementReceipt{}, err
	}
	if err := os.Rename(file.temp, file.name); err != nil {
		if !file.created {
			_ = os.Remove(file.backup)
		}
		return namespaceReplacementReceipt{}, err
	}
	file.replaced = true
	final, err := os.ReadFile(file.name)
	if err != nil {
		return namespaceReplacementReceipt{}, err
	}
	finalDigest := digestFileBytes(final)
	if finalDigest != file.tempDigest {
		return namespaceReplacementReceipt{}, fmt.Errorf("replacement changed staged bytes: %s", file.logical)
	}
	return namespaceReplacementReceipt{
		LogicalPath: file.logical, Primitive: "os.Rename",
		Contract: "same-directory-temp-over-destination-v1",
		SameDirectory: true, DestinationPreexisted: file.destinationPreexisted,
		TempDigest: file.tempDigest, ReplacementSuccess: true, FinalDigest: finalDigest,
	}, nil
}

func removeTransactionBackup(file transactionFile) error {
	if file.created {
		return nil
	}
	return os.Remove(file.backup)
}

func restoreTransaction(file transactionFile) {
	if file.created {
		if file.replaced {
			_ = os.Remove(file.name)
		}
		return
	}
	if file.replaced {
		_ = os.Remove(file.name)
		_ = os.Rename(file.backup, file.name)
		return
	}
	_ = os.Remove(file.backup)
}
