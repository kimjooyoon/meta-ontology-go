package linecaps

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAnalyzeDiscoversSortedFilesAndSkipsExcludedDirectories(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, filepath.Join("z", "last.go"), "package z\n")
	writeGoFile(t, root, filepath.Join("a", "first.go"), "package a\n")
	writeGoFile(t, root, filepath.Join("vendor", "ignored.go"), "package ignored\n"+strings.Repeat("\n", 10))
	writeGoFile(t, root, filepath.Join(".git", "ignored.go"), "package ignored\n"+strings.Repeat("\n", 10))
	files, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a/first.go", "z/last.go"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected discovery order: got %v want %v", files, want)
	}
	report, err := Analyze(root, nil, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatalf("discovered files should pass: %s", report.Text())
	}
}
func TestAnalyzeIsPermutationInvariantAndDoesNotMutatePaths(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "a.go", "package p\n\nfunc A() {\n\t_ = 1\n}\n")
	writeGoFile(t, root, "b.go", "package p\n\nfunc B() {\n\t_ = 2\n}\n")
	limits := Limits{MaxFileLines: 300, MaxFunctionLines: 2}
	absoluteA, err := filepath.Abs(filepath.Join(root, "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	firstPaths := []string{"b.go", absoluteA, "./a.go", "a.go"}
	secondPaths := []string{"a.go", "b.go"}
	firstBefore := append([]string(nil), firstPaths...)
	secondBefore := append([]string(nil), secondPaths...)
	first, err := Analyze(root, firstPaths, limits)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Analyze(root, secondPaths, limits)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) || first.Text() != second.Text() {
		t.Fatalf("path permutation changed result:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if !reflect.DeepEqual(firstPaths, firstBefore) || !reflect.DeepEqual(secondPaths, secondBefore) {
		t.Fatal("Analyze mutated its path inputs")
	}
}
