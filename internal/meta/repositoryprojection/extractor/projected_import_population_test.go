package extractor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProjectedCurrentCommandClosesActualImportPopulation(t *testing.T) {
	_, sourcePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("current source location unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourcePath), "../../../.."))
	logical := "cmd/gooo/body_codegen_tiny_test.go"
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(logical)))
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if err := projectedFinalConformance(root, logical, map[string][]byte{logical: source}); err != nil {
		t.Fatalf("unchanged current command did not close its projected imports: %#v", err)
	}
	t.Logf("current command source; full projected typecheck; elapsed=%s", time.Since(started))
}
