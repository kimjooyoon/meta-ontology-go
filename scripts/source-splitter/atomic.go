package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type stagedPart struct {
	temporary string
	target    string
}

func applySplit(plan splitPlan) error {
	staged := make([]stagedPart, len(plan.Parts))
	for index, part := range plan.Parts {
		temporary, err := stagePart(part, plan.Mode)
		if err != nil {
			for _, item := range staged {
				_ = os.Remove(item.temporary)
			}
			return err
		}
		staged[index] = stagedPart{temporary: temporary, target: part.Path}
	}
	defer func() {
		for _, item := range staged {
			_ = os.Remove(item.temporary)
		}
	}()
	for _, item := range staged[1:] {
		if _, err := os.Lstat(item.target); err == nil {
			return fmt.Errorf("split target already exists: %s", item.target)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	created := make([]string, 0, len(staged)-1)
	for index := 1; index < len(staged); index++ {
		if err := os.Rename(staged[index].temporary, staged[index].target); err != nil {
			for _, target := range created {
				_ = os.Remove(target)
			}
			return err
		}
		staged[index].temporary = ""
		created = append(created, staged[index].target)
	}
	if err := os.Rename(staged[0].temporary, staged[0].target); err != nil {
		for _, target := range created {
			_ = os.Remove(target)
		}
		return err
	}
	staged[0].temporary = ""
	directory, err := os.Open(filepath.Dir(staged[0].target))
	if err != nil {
		return err
	}
	err = directory.Sync()
	return errorsJoin(err, directory.Close())
}

func stagePart(part splitPart, mode os.FileMode) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(part.Path), ".source-split-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Chmod(mode); err == nil {
		_, err = file.Write(part.Data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(name)
		return "", errorsJoin(err, closeErr)
	}
	return name, nil
}

func errorsJoin(first, second error) error {
	if first != nil {
		return first
	}
	return second
}
