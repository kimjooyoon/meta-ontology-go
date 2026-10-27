package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestPackageGenerateRuleChangeKeepsConsumerSource(t *testing.T) {
	root, fixture := t.TempDir(), t.TempDir()
	writeWorkspaceSource(t, root, "go.mod", "module policyprobe\n\ngo 1.27.3\n")
	writeWorkspaceSource(t, root, "main.go", libraryEvolutionConsumer)
	if err := os.Mkdir(filepath.Join(root, "rules"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "examples", "runtime-package-surface")
	for _, name := range []string{"gooo.workspace.json", "contract.gooo", "inspect.gooo"} {
		raw, err := os.ReadFile(filepath.Join(path, name))
		if err != nil {
			t.Fatal(err)
		}
		writeWorkspaceSource(t, fixture, name, string(raw))
	}
	for revision := range 2 {
		if revision == 1 {
			addLibraryEvolutionRule(t, fixture)
		}
		generated := generateLibraryCommand(t, "rules", filepath.Join(fixture, "gooo.workspace.json"))
		writeWorkspaceSource(t, root, "rules/generated.go", string(generated))
		checkLibraryEvolutionConsumer(t, root, revision)
		consumer, err := os.ReadFile(filepath.Join(root, "main.go"))
		if err != nil || string(consumer) != libraryEvolutionConsumer {
			t.Fatal("rule change required handwritten consumer edits", err)
		}
	}
	t.Log("same-schema revisions=2; independent expected outcomes=6/6; handwritten consumer Go edits=0")
}

func addLibraryEvolutionRule(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "inspect.gooo")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const anchor = `    module + "/cmd/compare",`
	if strings.Count(string(raw), anchor) != 1 {
		t.Fatal("the independently specified source edit no longer has one target")
	}
	updated := strings.Replace(string(raw), anchor, anchor+"\n"+`    module + "/internal/recoverykernel",`, 1)
	writeWorkspaceSource(t, root, "inspect.gooo", updated)
}

func checkLibraryEvolutionConsumer(t *testing.T, root string, revision int) {
	t.Helper()
	inputs := libraryEvolutionInputs()
	raw, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "run", ".")
	command.Dir, command.Stdin = root, bytes.NewReader(raw)
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatal("unchanged Go consumer:", err, string(out))
	}
	var actual []libraryEvolutionAssessment
	if err := json.Unmarshal(out, &actual); err != nil || len(actual) != 3 {
		t.Fatal("consumer result population", err, string(out))
	}
	wants := libraryEvolutionExpectations[revision]
	for index, got := range actual {
		want := wants[index]
		if got.Status != want.Status || got.Missing != want.Missing || got.Undeclared != want.Undeclared ||
			got.Origin != inputs[index].Origin || !reflect.DeepEqual(got.Observed, inputs[index].Paths) {
			t.Fatalf("revision=%d case=%d got=%+v want=%+v", revision, index, got, want)
		}
	}
}
