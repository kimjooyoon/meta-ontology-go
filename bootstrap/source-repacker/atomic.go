package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func applyRepack(plan repackPlan) error {
	if len(plan.Edits) != 2 {
		return fmt.Errorf("repack requires exactly two edits")
	}
	staged := make([]stagedEdit, len(plan.Edits))
	for index, edit := range plan.Edits {
		temporary, err := stage(edit.Path, edit.After, os.FileMode(edit.Mode))
		if err != nil {
			for _, item := range staged {
				_ = os.Remove(item.temporary)
			}
			return err
		}
		staged[index] = stagedEdit{temporary: temporary, edit: edit}
	}
	defer func() {
		for _, item := range staged {
			_ = os.Remove(item.temporary)
		}
	}()
	if err := os.Rename(staged[1].temporary, staged[1].edit.Path); err != nil {
		return err
	}
	staged[1].temporary = ""
	if err := os.Rename(staged[0].temporary, staged[0].edit.Path); err != nil {
		_ = replaceFile(staged[1].edit.Path, staged[1].edit.Before, os.FileMode(staged[1].edit.Mode))
		return err
	}
	staged[0].temporary = ""
	directory, err := os.Open(filepath.Dir(staged[0].edit.Path))
	if err != nil {
		return err
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func stage(path string, data []byte, mode os.FileMode) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".source-repack-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Chmod(mode); err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(name)
		if err != nil {
			return "", err
		}
		return "", closeErr
	}
	return name, nil
}
