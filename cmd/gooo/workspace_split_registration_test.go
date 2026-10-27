package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const splitRegistryDirectory = "internal/meta/languagereadiness/languagesyntax"
const splitRegistryCorpus = "examples/language-syntax-roundtrip/corpus.json"
const splitWorkspaceDirectory = "examples/agent-tool-inventory"

func registeredSplitFixture(t *testing.T) workspaceSourceSplitOptions {
	t.Helper()
	root := t.TempDir()
	names, err := filepath.Glob(filepath.Join("../..", splitRegistryDirectory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if !strings.HasSuffix(name, "_test.go") {
			copyRegisteredSplitInput(t, root, filepath.Join(splitRegistryDirectory, filepath.Base(name)))
		}
	}
	contracts, err := filepath.Glob("../../internal/meta/generation/operation-input-contract*.gooo")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range contracts {
		copyRegisteredSplitInput(t, root, filepath.Join("internal/meta/generation", filepath.Base(name)))
	}
	for _, name := range []string{splitRegistryCorpus, splitWorkspaceDirectory + "/main.gooo",
		splitWorkspaceDirectory + "/gooo.workspace.json", splitWorkspaceDirectory + "/cases.json"} {
		copyRegisteredSplitInput(t, root, name)
	}
	return workspaceSourceSplitOptions{registryRoot: root, manifestPath: filepath.Join(root, splitWorkspaceDirectory, "gooo.workspace.json"),
		packagePath: "example/agent-tool-inventory", source: "main.gooo", before: "urn:gooo:agent-tool-inventory:inspect", into: "inspect.gooo"}
}

func copyRegisteredSplitInput(t *testing.T, root, name string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../..", name))
	if err != nil {
		t.Fatal(err)
	}
	writeRegisteredSplitInput(t, root, name, raw)
}

func writeRegisteredSplitInput(t *testing.T, root, name string, raw []byte) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, raw, 0o640); err != nil {
		t.Fatal(err)
	}
}

func saveRegisteredSplit(t *testing.T, options workspaceSourceSplitOptions) (workspaceSourceSplitOptions, workspaceSourceSplitReceipt) {
	t.Helper()
	proposal, err := prepareWorkspaceSourceSplit(options, OSFileReader{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(sourceSplitPublicReceipt(proposal))
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(options.registryRoot, "proposal.json")
	if err := os.WriteFile(name, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return workspaceSourceSplitOptions{manifestPath: options.manifestPath, applyReceipt: name}, proposal
}

func TestWorkspaceRegisteredSplitCLIInstallsAndExecutes(t *testing.T) {
	options := registeredSplitFixture(t)
	var stdout, stderr bytes.Buffer
	args := []string{"package", "split-source", "--package", options.packagePath, "--source", options.source,
		"--before", options.before, "--into", options.into, "--syntax-registry-root", options.registryRoot, "--json", options.manifestPath}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var proposal workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &proposal); err != nil {
		t.Fatal(err)
	}
	registration := proposal.Registration
	if registration == nil || len(registration.Outputs) != 2 || registration.Profile.Profile.Version != 5 || registration.Operation.ActivityID == "" ||
		!registration.Operation.UsedInputFact || !registration.Operation.GeneratedOutputFact {
		t.Fatal("joint source-owned proposal is incomplete")
	}
	writes, err := sourceSplitAllWrites(proposal)
	if err != nil || len(writes) != 5 {
		t.Fatal("joint publication must contain each source, manifest, corpus and native owner once", len(writes), err)
	}
	name := filepath.Join(options.registryRoot, "proposal.json")
	if err := os.WriteFile(name, stdout.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	args = []string{"package", "split-source", "--apply", name, "--json", options.manifestPath}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var applied workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &applied); err != nil || applied.Application == nil ||
		applied.Application.State != "OUTPUT_VERIFIED" || applied.Decision != "APPLIED" {
		t.Fatal("joint application was not verified", err)
	}
	stdout.Reset()
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(filepath.Dir(options.manifestPath), "cases.json"), options.manifestPath}, &stdout, &stderr); code != exitOK {
		t.Fatal("installed split program failed", code, stdout.String(), stderr.String())
	}
	var execution packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &execution); err != nil || execution.Result == nil ||
		execution.Result.Runtime.FinitePassed != 8 || execution.Result.Runtime.FiniteTotal != 8 ||
		!execution.Result.Runtime.RuntimeReplayed || execution.Result.Runtime.ModelCalls != 0 {
		t.Fatal("independent runtime oracle did not pass eight cases", err)
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("stale proposal was applied twice")
	}
	if err := verifyAppliedSourceSplit(applied); err != nil {
		t.Fatal("refused repeat changed installed files", err)
	}
}
