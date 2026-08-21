package protectedregions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalityViolationOrderAndNoMutation(t *testing.T) {
	before := []byte("package fixture\n" +
		"//gooo:protected:start id=\"fixture://z\"\n" +
		"const Z = 1\n" +
		"//gooo:protected:end id=\"fixture://z\"\n" +
		"//gooo:protected:start id=\"fixture://a\"\n" +
		"const A = 1\n" +
		"//gooo:protected:end id=\"fixture://a\"\n" +
		"//gooo:generated:start id=\"fixture://activity\" kind=\"activity\"\n" +
		"func Activity() int { return 1 }\n" +
		"//gooo:generated:end id=\"fixture://activity\" kind=\"activity\"\n")
	after := []byte(strings.Replace(strings.Replace(string(before), "const Z = 1", "const Z = 2", 1), "const A = 1", "const A = 2", 1))
	beforeCopy := append([]byte(nil), before...)
	afterCopy := append([]byte(nil), after...)
	report := ValidateLocality(before, after)
	if report.Valid() || len(report.Violations) != 2 {
		t.Fatalf("violations = %#v", report.Violations)
	}
	if report.Violations[0].ID != "fixture://a" || report.Violations[1].ID != "fixture://z" {
		t.Fatalf("violations were not sorted by stable ID: %#v", report.Violations)
	}
	if !bytes.Equal(before, beforeCopy) || !bytes.Equal(after, afterCopy) {
		t.Fatal("locality validation mutated its inputs")
	}
}
func TestLocalityRejectsGeneratedMarkerLineChanges(t *testing.T) {
	before := readFixture(t, "before.go")
	after := strings.Replace(string(before), "//gooo:generated:start id=\"fixture://activity\"", " //gooo:generated:start id=\"fixture://activity\"", 1)
	report := ValidateLocality(before, []byte(after))
	if !report.Before.Valid() || !report.After.Valid() || !hasLocalityIssue(report.Violations, LocalityUnownedChange) {
		t.Fatalf("generated marker line change was accepted without a locality violation: %#v", report)
	}
}
func hasIssue(issues []Issue, want IssueKind) bool {
	for _, issue := range issues {
		if issue.Kind == want {
			return true
		}
	}
	return false
}
func hasLocalityIssue(issues []LocalityIssue, want LocalityIssueKind) bool {
	for _, issue := range issues {
		if issue.Kind == want {
			return true
		}
	}
	return false
}
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "locality", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return source
}
