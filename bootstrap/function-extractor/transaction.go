package main

import (
	"fmt"
	"os"
	"sort"
)

type transactionFile struct {
	name   string
	temp   string
	backup string
}

func commitStaged(staged map[string]stagedFile) error {
	paths := make([]string, 0, len(staged))
	for path := range staged {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]transactionFile, 0, len(paths))
	for _, path := range paths {
		stage := staged[path]
		file := transactionFile{name: stage.name,
			temp: stage.name + ".extract.tmp", backup: stage.name + ".extract.bak"}
		if _, err := os.Stat(file.temp); !os.IsNotExist(err) {
			return fmt.Errorf("temporary extraction path exists: %s", file.temp)
		}
		if _, err := os.Stat(file.backup); !os.IsNotExist(err) {
			return fmt.Errorf("backup extraction path exists: %s", file.backup)
		}
		if err := os.WriteFile(file.temp, stage.data, os.FileMode(stage.mode)); err != nil {
			cleanupTransactions(files)
			return err
		}
		files = append(files, file)
	}
	committed := 0
	for index, file := range files {
		if err := os.Rename(file.name, file.backup); err != nil {
			rollbackTransactions(files, committed)
			return err
		}
		if err := os.Rename(file.temp, file.name); err != nil {
			_ = os.Rename(file.backup, file.name)
			rollbackTransactions(files, committed)
			return err
		}
		committed = index + 1
	}
	for _, file := range files {
		if err := os.Remove(file.backup); err != nil {
			return err
		}
	}
	return nil
}

func rollbackTransactions(files []transactionFile, committed int) {
	for index := committed - 1; index >= 0; index-- {
		_ = os.Remove(files[index].name)
		_ = os.Rename(files[index].backup, files[index].name)
	}
	cleanupTransactions(files)
}

func cleanupTransactions(files []transactionFile) {
	for _, file := range files {
		_ = os.Remove(file.temp)
	}
}
