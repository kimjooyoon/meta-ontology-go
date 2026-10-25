package syntaxregistration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The Go compiler independently evaluates the edited registry. This does not
// reuse the finite AST reader as the expected-value oracle.
func assertCompiledPartitionRegistry(t *testing.T, candidate PartitionRegistrationCandidate) {
	t.Helper()
	if err := compiledPartitionRegistry(t, candidate); err != nil {
		t.Fatal(err)
	}
}

func compiledPartitionRegistry(t *testing.T, candidate PartitionRegistrationCandidate) error {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	overlayPath, expected := compiledPartitionOverlay(t, root, t.TempDir(), candidate)
	command := exec.CommandContext(t.Context(), "go", "test", "-json", "-p", "1", "-overlay", overlayPath,
		"./internal/meta/languagereadiness/languagesyntax", "-run", "^TestCompiledPartitionRegistry$", "-count=1")
	command.Dir = root
	command.Env = append(os.Environ(), "GOOO_PARTITION_EXPECTED="+expected)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile registry: %w\n%s", err, output)
	}
	return checkCompiledPartitionRun(output)
}

func compiledPartitionOverlay(t *testing.T, root, temporary string, candidate PartitionRegistrationCandidate) (string, string) {
	t.Helper()
	replacements := map[string]string{}
	for name, raw := range candidate.Outputs {
		if filepath.Ext(name) != ".go" {
			continue
		}
		file := filepath.Join(temporary, filepath.Base(name))
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
		replacements[filepath.Join(root, filepath.FromSlash(name))] = file
	}
	expected := filepath.Join(temporary, "expected.json")
	if err := os.WriteFile(expected, candidate.Projection.Corpus, 0600); err != nil {
		t.Fatal(err)
	}
	harness := []byte(`package languagesyntax
import("encoding/json";"os";"reflect";"testing")
func TestCompiledPartitionRegistry(t *testing.T) {
 raw,err:=os.ReadFile(os.Getenv("GOOO_PARTITION_EXPECTED")); if err!=nil {t.Fatal(err)}
 var want Registry; if err:=json.Unmarshal(raw,&want);err!=nil {t.Fatal(err)}
 if got:=expectedRegistry();!reflect.DeepEqual(got,want) {t.Fatal("compiled registry differs from projected corpus")}
}
`)
	testFile := filepath.Join(temporary, "compiled_test.go")
	if err := os.WriteFile(testFile, harness, 0600); err != nil {
		t.Fatal(err)
	}
	replacements[filepath.Join(root, syntaxRoot, "zz_partition_compiled_test.go")] = testFile
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(temporary, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	return overlayPath, expected
}

func checkCompiledPartitionRun(output []byte) error {
	runs, passes := 0, 0
	for line := range bytes.SplitSeq(output, []byte{'\n'}) {
		var event struct{ Action, Test string }
		if json.Unmarshal(line, &event) != nil || event.Test != "TestCompiledPartitionRegistry" {
			continue
		}
		if event.Action == "run" {
			runs++
		}
		if event.Action == "pass" {
			passes++
		}
	}
	if runs != 1 || passes != 1 {
		return fmt.Errorf("compiled registry oracle did not execute exactly once: runs=%d passes=%d\n%s", runs, passes, output)
	}
	return nil
}
