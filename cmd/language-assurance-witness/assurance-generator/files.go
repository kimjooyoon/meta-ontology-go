package main

import (
	"bytes"
	"fmt"
	"os"
)

type outputFile struct {
	path string
	data []byte
}

func applyOutputs(outputs []outputFile, check bool) error {
	for _, output := range outputs {
		if check {
			current, err := os.ReadFile(output.path)
			if err != nil {
				return err
			}
			if !bytes.Equal(current, output.data) {
				return fmt.Errorf("%s is stale", output.path)
			}
			continue
		}
		if err := os.WriteFile(output.path, output.data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
