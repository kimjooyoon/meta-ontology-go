package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitPartPathPreservesBuildDomain(t *testing.T) {
	tests := map[string]string{
		"fixture.go":                  "fixture_split02.go",
		"fixture_test.go":             "fixture_split02_test.go",
		"fixture_linux.go":            "fixture_split02_linux.go",
		"fixture_linux_amd64_test.go": "fixture_split02_linux_amd64_test.go",
	}
	for input, expected := range tests {
		actual, err := splitPartPath(input, 2)
		if err != nil || actual != expected {
			t.Fatalf("splitPartPath(%q) = %q, %v; want %q", input, actual, err, expected)
		}
	}
}

func TestPlanSourceSplitsDeclarationsAndImports(t *testing.T) {
	root := t.TempDir()
	subject := "fixture_linux_test.go"
	source := `//go:build linux

package fixture

import "fmt"

func first() string {
	return fmt.Sprint("a")
}

func second() int {
	return 2
}
`
	if err := os.WriteFile(filepath.Join(root, subject), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := planSource(root, subject, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parts) != 2 || !strings.HasSuffix(plan.Parts[1].Subject, "_split02_linux_test.go") {
		t.Fatalf("unexpected parts: %#v", plan.Parts)
	}
	for _, part := range plan.Parts {
		if physicalLines(part.Data) > 10 || !strings.Contains(string(part.Data), "//go:build linux") {
			t.Fatalf("invalid split part %s:\n%s", part.Subject, part.Data)
		}
	}
	for _, part := range plan.Parts {
		data := string(part.Data)
		if strings.Contains(data, `"fmt"`) != strings.Contains(data, "fmt.Sprint") {
			t.Fatalf("import usage mismatch:\n%s", part.Data)
		}
	}
}
