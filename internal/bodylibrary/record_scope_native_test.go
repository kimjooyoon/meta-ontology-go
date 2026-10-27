package bodylibrary

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLibraryRecordScopeKeepsOneNativeConsumerAcrossUnusedTypeChange(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "scope"), 0700); err != nil {
		t.Fatal(err)
	}
	consumer := `package consumer
import("testing"; "example.test/recordscope/scope")
func TestValues(t *testing.T) {
 for _,c:=range []struct{input,want int64}{{-2,1},{0,1},{7,22}} {
  got:=scope.Evaluate(scope.R_Request{F_value:c.input})
  if got.F_value!=c.want {t.Fatalf("input %d: got %d want %d",c.input,got.F_value,c.want)}
 }
}
`
	for name, raw := range map[string]string{"go.mod": "module example.test/recordscope\ngo 1.27.3\n", "consumer_test.go": consumer} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	source := recordScopeFixture(t)
	variants := [][]byte{source, append(append([]byte(nil), source...), []byte("\nentity LaterUnused id \"urn:gooo:record-scope:later\" fields {\nfield value id \"urn:gooo:record-scope:later:value\" type integer required one\n}\n")...)}
	for index, raw := range variants {
		generated, err := Generate(context.Background(), "scope.gooo", raw, "Evaluate", "scope")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(generated, fmt.Sprintf("// Lowered Gooo source: sha256:%x", sha256.Sum256(raw))) {
			t.Fatal("full source identity was pruned")
		}
		if err := os.WriteFile(filepath.Join(root, "scope", "semantic.gooo.go"), []byte(generated), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./...")
		command.Dir = root
		command.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
		output, err := command.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("variant %d: %v\n%s", index, err, output)
		}
		stored, err := os.ReadFile(filepath.Join(root, "consumer_test.go"))
		if err != nil || string(stored) != consumer {
			t.Fatal("native consumer changed", err)
		}
	}
}

func TestLibraryRecordScopeDoesNotHideInvalidUnusedLayout(t *testing.T) {
	source := recordScopeFixture(t)
	invalid := "\nentity InvalidUnused id \"urn:gooo:record-scope:invalid\" fields {\n"
	for i := range 33 {
		invalid += fmt.Sprintf("field f%d id \"urn:gooo:record-scope:invalid:%d\" type integer required one\n", i, i)
	}
	generated, err := Generate(context.Background(), "scope.gooo", append(source, []byte(invalid+"}\n")...), "Evaluate", "scope")
	if generated != "" || err == nil || !strings.Contains(err.Error(), "declares 33 fields") {
		t.Fatal("unreachable source escaped layout validation", generated, err)
	}
}
