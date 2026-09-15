package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func readMetadataAt(path string) (Metadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Metadata{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Metadata{}, fmt.Errorf("metadata is not a regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var metadata Metadata
	if err := decoder.Decode(&metadata); err != nil {
		return Metadata{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Metadata{}, fmt.Errorf("multiple JSON values")
		}
		return Metadata{}, err
	}
	return metadata, nil
}
func readDataFile(path string, maxSize int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("projection is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if maxSize > 0 {
		data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxSize {
			return nil, ErrEntryTooLarge
		}
		return data, nil
	}
	return io.ReadAll(file)
}
