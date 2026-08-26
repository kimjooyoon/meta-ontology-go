package main

import (
	"os"
	"testing"
)

func snapshotFile(t *testing.T, filename string) ([]byte, os.FileInfo) {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	return data, info
}
