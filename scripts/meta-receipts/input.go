package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func decodeJSON(path string, target any) error {
	input, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer input.Close()
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("decode %s: expected one JSON value", path)
	}
	return nil
}
