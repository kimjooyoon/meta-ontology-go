package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func installTransaction(file *transactionFile) (namespaceReplacementReceipt, error) {
	if !sameDirectory(file.name, file.temp) {
		return namespaceReplacementReceipt{}, fmt.Errorf("replacement paths are not same-directory: %s", file.logical)
	}
	if runtime.GOOS != "linux" {
		return namespaceReplacementReceipt{}, fmt.Errorf("namespace replacement unsupported on GOOS %s: %s", runtime.GOOS, file.logical)
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
	file.tempCreated = false
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
		Contract:      linuxNamespaceReplacementContract,
		GOOS:          runtime.GOOS,
		SameDirectory: true, DestinationPreexisted: file.destinationPreexisted,
		TempDigest: file.tempDigest, ReplacementSuccess: true, FinalDigest: finalDigest,
	}, nil
}

func removeTransactionBackup(file transactionFile) error {
	if file.created || !file.backupCreated {
		return nil
	}
	return os.Remove(file.backup)
}

func restoreTransaction(file transactionFile) error {
	if file.created {
		if !file.replaced {
			return nil
		}
		if err := os.Remove(file.name); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if file.replaced {
		if err := os.Rename(file.backup, file.name); err != nil {
			if restoreErr := restoreOriginal(file); restoreErr != nil {
				return fmt.Errorf("restore backup: %w; restore original: %v", err, restoreErr)
			}
			return removeBackupDuringRollback(file)
		}
		file.backupCreated = false
		return nil
	}
	return removeBackupDuringRollback(file)
}

func removeBackupDuringRollback(file transactionFile) error {
	if file.created || !file.backupCreated {
		return nil
	}
	err := os.Remove(file.backup)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func restoreOriginal(file transactionFile) error {
	if file.original == nil {
		return fmt.Errorf("original bytes are unavailable for %s", file.logical)
	}
	restored, err := os.CreateTemp(filepath.Dir(file.name), ".extract-restore-*")
	if err != nil {
		return err
	}
	path := restored.Name()
	defer os.Remove(path)
	if err := restored.Chmod(file.originalMode); err != nil {
		_ = restored.Close()
		return err
	}
	if _, err := restored.Write(file.original); err != nil {
		_ = restored.Close()
		return err
	}
	if err := restored.Close(); err != nil {
		return err
	}
	return os.Rename(path, file.name)
}
