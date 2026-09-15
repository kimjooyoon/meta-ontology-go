package shadow

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type productionProcess struct {
	root string
	bin  string
}

func buildProductionProcess(t *testing.T) productionProcess {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate equivalence test")
	}
	root := filepath.Dir(file)
	for range 4 {
		root = filepath.Dir(root)
	}
	bin := filepath.Join(t.TempDir(), "gooo")
	command := exec.Command("go", "build", "-o", bin, "./cmd/gooo")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build production command: %v\n%s", err, output)
	}
	return productionProcess{root: root, bin: bin}
}
func (p productionProcess) invoke(t *testing.T, fixture productionFixture) productionOutput {
	t.Helper()
	directory := t.TempDir()
	paths := map[string]string{}
	for name, data := range fixture.files {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		paths[name] = path
	}
	args := []string{"selective-ci", "shadow", "--base-snapshot", paths["base_snapshot.json"], "--head-snapshot", paths["head_snapshot.json"], "--plan-input", paths["plan_input.json"], "--evidence-input", paths["evidence_input.json"], "--lane-input", paths["lane_input.json"]}
	command := exec.Command(p.bin, args...)
	command.Dir = p.root
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("production command: %v\nstderr=%s\nstdout=%s", err, stderr.Bytes(), stdout.Bytes())
	}
	if stderr.Len() != 0 {
		t.Fatalf("production command stderr = %q", stderr.String())
	}
	var output productionOutput
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&output); err != nil {
		t.Fatalf("decode production receipt: %v\n%s", err, stdout.Bytes())
	}
	return output
}
