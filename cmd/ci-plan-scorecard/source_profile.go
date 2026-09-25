package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/ciplanusecase"
)

func sourceProfile(root string) (ciplanusecase.SourceProfile, error) {
	profile := ciplanusecase.SourceProfile{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".gooo" && extension != ".go" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := bytes.Count(raw, []byte{'\n'})
		if len(raw) != 0 && raw[len(raw)-1] != '\n' {
			lines++
		}
		if extension == ".gooo" {
			profile.GoooFiles++
			profile.GoooLines += lines
		} else {
			profile.GoFiles++
			profile.GoLines += lines
		}
		return nil
	})
	return profile, err
}
