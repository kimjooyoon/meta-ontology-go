package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceSourceSplitCLIProposalCanBeLoaded(t *testing.T) {
	root := t.TempDir()
	inputs := map[string][]byte{}
	for _, name := range []string{"main.gooo", "workspace.json"} {
		raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "named-inputs", name))
		if err != nil {
			t.Fatal(err)
		}
		inputs[name] = raw
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	args := []string{"package", "split-source", "--package", "named/inputs", "--source", "main.gooo",
		"--before", "urn:gooo:example:named-inputs:net", "--into", "net.gooo", "--json", filepath.Join(root, "workspace.json")}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stdout.String(), stderr.String())
	}
	var receipt workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "PROPOSED" || receipt.Result == nil || receipt.UpdatedWorkspace == nil || receipt.ManifestDigest == "" {
		t.Fatal("incomplete split proposal")
	}
	for name, original := range inputs {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || !bytes.Equal(raw, original) {
			t.Fatal("proposal mutated workspace", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "net.gooo")); !os.IsNotExist(err) {
		t.Fatal("proposal wrote the destination", err)
	}
	for _, file := range receipt.Result.Files {
		if err := os.WriteFile(filepath.Join(root, file.Filename), []byte(file.Source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(receipt.UpdatedWorkspace)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"package", "resolve", "--json", filepath.Join(root, "workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal("installed proposal does not load through the actual consumer", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("split again must reject the occupied destination", code)
	}
	var refused workspaceSourceSplitReceipt
	if err := json.Unmarshal(stdout.Bytes(), &refused); err != nil || refused.Decision != "FAIL_CLOSED" ||
		refused.Result != nil || refused.UpdatedWorkspace != nil {
		t.Fatal("refused split leaked a partial proposal", refused, err)
	}
}
