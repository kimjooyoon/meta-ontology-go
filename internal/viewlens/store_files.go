package viewlens

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/originpack"
)

func withProfileLock(path string, work func() error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	locked, err := tryLock(file)
	if err != nil {
		return err
	}
	if !locked {
		return fmt.Errorf("PROFILE_BUSY")
	}
	return work()
}

func writeProfile(path string, raw []byte) (int, error) {
	packed, err := originpack.Pack(raw)
	if err != nil {
		return 0, err
	}
	file, err := preparedFile(filepath.Dir(path), packed)
	if err != nil {
		return 0, err
	}
	defer os.Remove(file)
	return len(packed), os.Rename(file, path)
}

func preparedFile(directory string, packed []byte) (string, error) {
	file, err := os.CreateTemp(directory, ".view-lens-")
	if err != nil {
		return "", err
	}
	path := file.Name()
	complete := false
	defer func() {
		file.Close()
		if !complete {
			os.Remove(path)
		}
	}()
	if _, err := file.Write(packed); err != nil {
		return "", err
	}
	if err := file.Sync(); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	complete = true
	return path, nil
}

func archiveEvent(path string, raw []byte) (int, error) {
	if previous, size, err := readPacked(path); !os.IsNotExist(err) {
		if err != nil || !bytes.Equal(previous, raw) {
			return 0, fmt.Errorf("PROFILE_ARCHIVE_CONFLICT")
		}
		return size, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return 0, err
	}
	packed, err := originpack.Pack(raw)
	if err != nil {
		return 0, err
	}
	staged, err := preparedFile(filepath.Dir(path), packed)
	if err != nil {
		return 0, err
	}
	defer os.Remove(staged)
	if err := os.Link(staged, path); err != nil {
		if !os.IsExist(err) {
			return 0, err
		}
		old, size, err := readPacked(path)
		if err != nil || !bytes.Equal(old, raw) {
			return 0, fmt.Errorf("PROFILE_ARCHIVE_CONFLICT")
		}
		return size, nil
	}
	return len(packed), nil
}
