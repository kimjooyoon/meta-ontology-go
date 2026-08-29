package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type transactionFile struct {
	logical               string
	name                  string
	temp                  string
	backup                string
	created               bool
	destinationPreexisted bool
	tempDigest            string
	replaced              bool
}

type stagedTransaction struct {
	files    []transactionFile
	receipts []namespaceReplacementReceipt
}

const linuxNamespaceReplacementContract = "same-directory-temp-over-destination/linux-v1"

func commitStaged(staged map[string]stagedFile) (stagedTransaction, error) {
	paths := make([]string, 0, len(staged))
	for path := range staged {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]transactionFile, 0, len(paths))
	cleanupAndFail := func(err error) (stagedTransaction, error) {
		cleanupTransactions(files)
		return stagedTransaction{}, err
	}
	for _, path := range paths {
		stage := staged[path]
		file := transactionFile{logical: path, name: stage.name,
			temp: stage.name + ".extract.tmp", backup: stage.name + ".extract.bak", created: stage.created}
		files = append(files, file)
		if _, err := os.Lstat(file.temp); !os.IsNotExist(err) {
			return cleanupAndFail(fmt.Errorf("temporary extraction path exists: %s", file.temp))
		}
		if _, err := os.Lstat(file.backup); !os.IsNotExist(err) {
			return cleanupAndFail(fmt.Errorf("backup extraction path exists: %s", file.backup))
		}
		if !file.created {
			info, err := os.Lstat(file.name)
			if err != nil {
				return cleanupAndFail(err)
			}
			if !info.Mode().IsRegular() {
				return cleanupAndFail(fmt.Errorf("replacement target is not a regular file: %s", file.logical))
			}
			file.destinationPreexisted = true
		} else if _, err := os.Lstat(file.name); err == nil || !os.IsNotExist(err) {
			return cleanupAndFail(fmt.Errorf("creation target exists: %s", file.logical))
		}
		if err := os.WriteFile(file.temp, stage.data, os.FileMode(stage.mode)); err != nil {
			return cleanupAndFail(err)
		}
		temp, err := os.ReadFile(file.temp)
		if err != nil {
			return cleanupAndFail(err)
		}
		file.tempDigest = digestFileBytes(temp)
		files[len(files)-1] = file
	}
	receipts := make([]namespaceReplacementReceipt, 0, len(files))
	for index := range files {
		receipt, err := installTransaction(&files[index])
		if err != nil {
			rollbackTransactions(files, index+1)
			return stagedTransaction{}, err
		}
		receipts = append(receipts, receipt)
	}
	return stagedTransaction{files: files, receipts: receipts}, nil
}

func removeTransactionBackups(files []transactionFile) error {
	for index := range files {
		if err := removeTransactionBackup(files[index]); err != nil {
			return err
		}
	}
	return nil
}

func rollbackTransactions(files []transactionFile, committed int) {
	for index := committed - 1; index >= 0; index-- {
		restoreTransaction(files[index])
	}
	cleanupTransactions(files)
}

func cleanupTransactions(files []transactionFile) {
	for _, file := range files {
		_ = os.Remove(file.temp)
	}
}

func preserveDestination(file *transactionFile) error {
	data, err := os.ReadFile(file.name)
	if err != nil {
		return err
	}
	info, err := os.Stat(file.name)
	if err != nil {
		return err
	}
	backup, err := os.OpenFile(file.backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := backup.Write(data); err != nil {
		_ = backup.Close()
		_ = os.Remove(file.backup)
		return err
	}
	if err := backup.Close(); err != nil {
		_ = os.Remove(file.backup)
		return err
	}
	return nil
}

func sameDirectory(left, right string) bool {
	return filepath.Clean(filepath.Dir(left)) == filepath.Clean(filepath.Dir(right))
}

func digestFileBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
