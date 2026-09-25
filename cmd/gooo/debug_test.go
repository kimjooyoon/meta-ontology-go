package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDebugRequiresEntryAndSource(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := runDebug(nil, &stdout, &stderr); code != exitUsage {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(stderr.String(), "usage: gooo debug") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDebugRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := runDebug([]string{"--unknown"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("code = %d", code)
	}
}
