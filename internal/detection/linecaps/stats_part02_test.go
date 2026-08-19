package linecaps

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeLineMetricsTextIncludesPerLanguageRows(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "a.go", "package a\n")
	writeMetricFile(t, root, "b.gooo", "intent: b\n")
	writeMetricFile(t, root, "c.txt", "x\n")
	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	text := report.Text()
	if !strings.Contains(text, "language totals: go_files=1 gooo_files=1 go_lines=2 gooo_lines=2") {
		t.Fatalf("text missing language totals:\n%s", text)
	}
	if !strings.Contains(text, "go files: count=1 lines=2") || !strings.Contains(text, "a.go\tlines=2") {
		t.Fatalf("text missing go file rows:\n%s", text)
	}
	if !strings.Contains(text, "gooo files: count=1 lines=2") || !strings.Contains(text, "b.gooo\tlines=2") {
		t.Fatalf("text missing gooo file rows:\n%s", text)
	}
}
func TestAnalyzeLineMetricsRejectsEmptyRoot(t *testing.T) {
	if _, err := AnalyzeLineMetrics(""); err == nil {
		t.Fatal("empty root was accepted")
	}
}
func directoryForPath(report LineMetricsReport, path string) DirectoryMetric {
	for _, directory := range report.Directories {
		if directory.Path == path {
			return directory
		}
	}
	return DirectoryMetric{}
}
func ensureMetricDir(t *testing.T, root, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, path), 0o755); err != nil {
		t.Fatal(err)
	}
}
func writeMetricFile(t *testing.T, root, path, source string) {
	t.Helper()
	ensureMetricDir(t, root, filepath.Dir(path))
	if err := os.WriteFile(filepath.Join(root, path), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}
