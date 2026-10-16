package main

import (
	"fmt"
	"io"
	"os"
)

func readBounded(path string, maximum int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if maximum < 0 || !info.Mode().IsRegular() || info.Size() > maximum {
		return nil, fmt.Errorf("input exceeds its declared regular-file budget: %s", path)
	}
	raw, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maximum {
		return nil, fmt.Errorf("input grew beyond its declared byte budget: %s", path)
	}
	return raw, nil
}
