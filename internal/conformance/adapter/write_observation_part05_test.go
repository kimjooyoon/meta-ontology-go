package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newObserverWithTempSetup(t *testing.T, request Request, setup func(string)) *NoWriteObserver {
	t.Helper()
	root := t.TempDir()
	tempRoot := filepath.Join(root, "tmp")
	if err := os.Mkdir(tempRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(root, "source.gooo")
	outputPath := filepath.Join(root, "output.go")
	if err := os.WriteFile(sourcePath, []byte("entity billing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte("package billing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeTemp(t, filepath.Join(tempRoot, "stable.tmp"), "stable")
	if setup != nil {
		setup(tempRoot)
	}
	observer, err := NewNoWriteObserver(requestObservationBinding(request), ObserverPaths{
		SourcePath: sourcePath, OutputPath: outputPath, TempRoot: tempRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	attachVerifiedWorkflow(t, observer, request)
	attachVerifiedMutation(t, observer, request)
	return observer
}
func writeTemp(t *testing.T, path, value string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
}
func removeTemp(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
func renameTemp(t *testing.T, oldPath, newPath string) {
	t.Helper()
	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatal(err)
	}
}
func assertOracleFailure(t *testing.T, evaluation Evaluation, code string) {
	t.Helper()
	if evaluation.Matched || evaluation.OracleCode != code || evaluation.PromotionEligible {
		t.Fatalf("expected %s non-promotion failure, got %+v", code, evaluation)
	}
	if !strings.Contains(evaluation.Detail, code) {
		t.Fatalf("failure detail omitted %s: %q", code, evaluation.Detail)
	}
}
