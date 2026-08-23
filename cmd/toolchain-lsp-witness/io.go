package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func readStrict[T any](path string) (T, error) {
	var value T
	file, err := os.Open(path)
	if err != nil { return value, err }
	defer file.Close()
	decoder := json.NewDecoder(file); decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil { return value, err }
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) { return value, fmt.Errorf("trailing JSON") }
	return value, nil
}

func writeExclusive(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil { return err }
	raw = append(raw, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil { return err }
	if _, err = file.Write(raw); err != nil { file.Close(); return err }
	return file.Close()
}
