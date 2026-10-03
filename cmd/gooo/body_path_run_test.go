package main

import (
	"bytes"
	"testing"
)

func TestBodyPathRunCLIEntry(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"body-path-run", "--help"}, &stdout, &stderr); code != exitOK ||
		!bytes.Contains(stderr.Bytes(), []byte("gooo body-path-run")) || stdout.Len() != 0 {
		t.Fatal("missing CLI entry", code, stderr.String())
	}
}
