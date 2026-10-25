package sourcerepack

import (
	"fmt"
	"os"
	"path/filepath"
)

func applyRepack(plan repackPlan) error {
	return applyRepackWithTemp(plan, os.CreateTemp)
}

func applyRepackWithTemp(plan repackPlan, create tempFileCreator) error {
	if len(plan.Edits) != 2 {
		return fmt.Errorf("repack requires exactly two edits")
	}
	staged, err := prepareRepackEdits(plan, create)
	if err != nil {
		return err
	}
	defer removeStagedEdits(staged)
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
