package toolchainrelease

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelativeV3BuildRejectionPrecedesModelAndSourceAccess(t *testing.T) {
	work := t.TempDir()
	missing := filepath.Join(work, "missing-gooo")
	input := BuildInput{Root: filepath.Join(work, "missing-source")}
	if err := smokeRelativeV3(missing, work, input); err == nil {
		t.Fatal("unobserved executable accepted")
	}
	entries, err := os.ReadDir(work)
	if err != nil || len(entries) != 0 {
		t.Fatalf("unobserved executable created profile artifacts: %v, %v", entries, err)
	}
}
