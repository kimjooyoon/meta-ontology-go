package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

const header = "package sample\nnamespace sample\n"
const scalar = "entity Integer id \"urn:gooo:type:integer\"\n"
const body = "activity Double(Integer) -> Integer id \"urn:sample:double\" computes `return input * 2`\n"

func TestBundleProjectionUsesOnlySelectedFamily(t *testing.T) {
	files := fstest.MapFS{
		"source.gooo":       {Data: []byte(header + scalar)},
		"source_body.gooo":  {Data: []byte(header + body)},
		"unrelated.gooo":    {Data: []byte("broken source")},
		"nested/other.gooo": {Data: []byte("not a member")},
	}
	got, err := bundleProjection(files, "source*.gooo", "Double")
	if err != nil || !strings.Contains(got, "func Double(") || !strings.Contains(got, "urn:sample:double") {
		t.Fatal("explicit family not generated", err, got)
	}
	for _, pattern := range []string{"absent*.gooo", "[", "*.gooo", "source.gooo"} {
		if output, err := bundleProjection(files, pattern, "Double"); err == nil || output != "" {
			t.Fatal("invalid source selection produced output", pattern, err)
		}
	}
}

func TestLegacyAbsolutePathAndFailureLeaveOutputIntact(t *testing.T) {
	dir := t.TempDir()
	source, output := filepath.Join(dir, "source.gooo"), filepath.Join(dir, "output.go")
	if err := os.WriteFile(source, []byte(header+scalar+body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := generateArgs([]string{source, output, "Double"}); err != nil {
		t.Fatal("legacy absolute path rejected", err)
	}
	before, err := os.ReadFile(output)
	if err != nil || !strings.Contains(string(before), "func Double(") {
		t.Fatal(err)
	}
	if err := generateArgs([]string{source, output, "Missing"}); err == nil {
		t.Fatal("missing activity accepted")
	}
	after, err := os.ReadFile(output)
	if err != nil || string(after) != string(before) {
		t.Fatal("failed projection changed output", err)
	}
}
