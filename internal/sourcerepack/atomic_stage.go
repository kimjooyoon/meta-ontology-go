package sourcerepack

import (
	"os"
	"path/filepath"
)

type tempFileCreator func(string, string) (*os.File, error)

func prepareRepackEdits(plan repackPlan, create tempFileCreator) ([]stagedEdit, error) {
	staged := make([]stagedEdit, len(plan.Edits))
	for index, edit := range plan.Edits {
		temporary, err := stageWithTemp(edit.Path, edit.After, os.FileMode(edit.Mode), create)
		if err != nil {
			removeStagedEdits(staged)
			return nil, err
		}
		staged[index] = stagedEdit{temporary: temporary, edit: edit}
	}
	return staged, nil
}

func removeStagedEdits(staged []stagedEdit) {
	for _, item := range staged {
		if item.temporary != "" {
			_ = os.Remove(item.temporary)
		}
	}
}

func stage(path string, data []byte, mode os.FileMode) (string, error) {
	return stageWithTemp(path, data, mode, os.CreateTemp)
}

func stageWithTemp(path string, data []byte, mode os.FileMode, create tempFileCreator) (string, error) {
	file, err := create(filepath.Dir(path), ".source-repack-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err = file.Chmod(mode); err == nil {
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
