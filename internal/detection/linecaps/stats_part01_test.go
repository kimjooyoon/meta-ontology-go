package linecaps

import (
	"path/filepath"
	"testing"
)

func TestAnalyzeLineMetricsReturnsGoAndGooLangLineCounts(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, "a.go", "package a\n\nvar X = 1\n")
	writeMetricFile(t, root, "b.gooo", "intent: a\n\nnode: B\n")
	writeMetricFile(t, root, filepath.Join("sub", "c.go"), "package sub\n\nfunc C() {}\n")
	writeMetricFile(t, root, filepath.Join("sub", "d.gooo"), "intent: sub\n")
	writeMetricFile(t, root, filepath.Join("sub", "nested", "e.gooo"), "intent: nested\n")
	writeMetricFile(t, root, filepath.Join("sub", "nested", "f.go"), "package nested\n")
	writeMetricFile(t, root, filepath.Join("vendor", "ignore.go"), "package ignored\n")
	writeMetricFile(t, root, filepath.Join(".git", "ignore.gooo"), "intent: ignored\n")

	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Root != filepath.ToSlash(root) {
		t.Fatalf("unexpected root: %s", report.Root)
	}
	total := report.Total()
	if total.GoFiles != 3 {
		t.Fatalf("unexpected go file count: %#v", total)
	}
	if total.GoooFiles != 3 {
		t.Fatalf("unexpected gooo file count: %#v", total)
	}
	if total.GoLines != 7 {
		t.Fatalf("unexpected go lines: %#v", total.GoLines)
	}
	if total.GoooLines != 5 {
		t.Fatalf("unexpected gooo lines: %#v", total.GoooLines)
	}
	if total.DirectFiles != 2 || total.RecursiveFiles != 6 {
		t.Fatalf("unexpected file count: %#v", total)
	}
	if total.DirectFolders != 1 || total.RecursiveFolders != 2 {
		t.Fatalf("unexpected folder count: %#v", total)
	}

	sub := directoryForPath(report, "sub")
	if sub.DirectFiles != 2 || sub.RecursiveFiles != 4 {
		t.Fatalf("unexpected nested directory file count: %#v", sub)
	}
	if sub.DirectFolders != 1 || sub.RecursiveFolders != 1 {
		t.Fatalf("unexpected nested directory folder count: %#v", sub)
	}
}
func TestAnalyzeLineMetricsIncludesNonGoSourceLines(t *testing.T) {
	root := t.TempDir()
	writeMetricFile(t, root, filepath.Join("readme.md"), "alpha\nbeta\n")

	report, err := AnalyzeLineMetrics(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 || report.Files[0].Language != FileLanguageOther {
		t.Fatalf("unexpected file language metric: %#v", report.Files)
	}
}
