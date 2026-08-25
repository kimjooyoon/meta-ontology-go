package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTopologyFailuresExposeRelativePhysicalSubject(t *testing.T) {
	root := t.TempDir()
	dense := filepath.Join(root, "dense")
	if err := os.MkdirAll(dense, 0o755); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 11; index++ {
		name := filepath.Join(dense, fmt.Sprintf("entry-%02d", index))
		if err := os.WriteFile(name, []byte("evidence\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := topologyFailures(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Direct != 1 || result.Mixed != 0 || len(result.Subjects) != 1 {
		t.Fatalf("topology = %#v", result)
	}
	subject := result.Subjects[0]
	if subject.Indicator != directIndicator || subject.Physical != "dense" ||
		subject.Value != 11 || subject.Limit != 10 || subject.Consumer != "radix-sharder" ||
		subject.Operation != "split-object-bucket" {
		t.Fatalf("subject = %#v", subject)
	}
}

func TestTopologyFailuresExcludeProjectRoot(t *testing.T) {
	root := t.TempDir()
	for index := 0; index < 11; index++ {
		name := filepath.Join(root, fmt.Sprintf("root-entry-%02d", index))
		if err := os.WriteFile(name, []byte("root evidence\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	result, err := topologyFailures(root)
	if err != nil {
		t.Fatal(err)
	}
	if result.Direct != 0 || result.Mixed != 0 || len(result.Subjects) != 0 {
		t.Fatalf("project root must remain exempt: %#v", result)
	}
}

func TestBlockingErrorNamesPhysicalSubject(t *testing.T) {
	report := evidence{
		Indicators: []indicator{{ID: directIndicator, Value: 1, Limit: 0, Blocking: true}},
		Subjects:   []subject{{Indicator: directIndicator, Physical: ".github/workflows", Value: 42, Limit: 10}},
	}
	err := requireBlockingZero(report)
	if err == nil || !strings.Contains(err.Error(), "subjects=.github/workflows(42>10)") {
		t.Fatalf("error = %v", err)
	}
}
