package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func readJSON[T any](name string) (T, []byte, error) {
	var value T
	data, err := os.ReadFile(name)
	if err != nil {
		return value, nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&value); err != nil {
		return value, nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return value, nil, fmt.Errorf("%s has trailing JSON", name)
	}
	return value, data, nil
}

func writeExclusive(name string, data []byte) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
