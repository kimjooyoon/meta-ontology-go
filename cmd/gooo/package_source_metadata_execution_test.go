package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func copySourceMetadataWorkspace(t *testing.T, original string) string {
	t.Helper()
	raw, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]json.RawMessage
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	var packages []map[string]json.RawMessage
	if err = json.Unmarshal(manifest["packages"], &packages); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, pkg := range packages {
		delete(pkg, "name")
		delete(pkg, "imports")
		var sources []string
		if err = json.Unmarshal(pkg["sources"], &sources); err != nil {
			t.Fatal(err)
		}
		for _, name := range sources {
			content, readErr := os.ReadFile(filepath.Join(filepath.Dir(original), name))
			if readErr != nil {
				t.Fatal(readErr)
			}
			if err = os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0700); err != nil {
				t.Fatal(err)
			}
			writeWorkspaceSource(t, root, name, string(content))
		}
	}
	manifest["packages"], err = json.Marshal(packages)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return writeWorkspaceManifest(t, root, string(raw))
}

func TestWorkspaceSourceMetadataExecutesAndReplaysImportedCalls(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	manifest := copySourceMetadataWorkspace(t, "../../examples/package-body-calls/gooo.workspace.json")
	cases := "../../examples/package-body-calls/cases.json"
	first, raw := runSavedPackageFixture(t, "execute", "--json", "--cases", cases, manifest)
	if first.Decision != "PASS" || first.Result.Runtime.FinitePassed != 8 || first.Result.Runtime.FiniteTotal != 8 ||
		len(first.Result.Program.PureCalls.Sites) != 2 || len(first.Result.Program.Activities) != 2 {
		t.Fatal("source metadata did not preserve imported execution", first.Decision)
	}
	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	moved := copySourceMetadataWorkspace(t, manifest)
	replay, _ := runSavedPackageFixture(t, "replay", "--json", "--receipt", path, "--cases", cases, moved)
	if replay.Result.Replay == nil || replay.Result.Replay.ModelCalls != 0 || replay.Result.Runtime.FinitePassed != 8 ||
		replay.Result.Runtime.FiniteTotal != 8 || replay.Result.Runtime.ModelCalls != 0 || replay.ReplayedFrom == "" {
		t.Fatal("source metadata did not preserve moved saved execution")
	}
}

func TestWorkspaceSourceMetadataResumesWithGoooPolicy(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	manifest := copySourceMetadataWorkspace(t, "../../examples/called-body-construction/gooo.workspace.json")
	checkpoint := copySourceMetadataWorkspace(t, "../../examples/package-assembly-policy/checkpoint.workspace.json")
	policy := copySourceMetadataWorkspace(t, "../../examples/package-assembly-policy/gooo.workspace.json")
	cases := "../../examples/called-body-construction/cases.json"
	first, raw := runSavedPackageFixture(t, "execute", "--json", "--cases", cases,
		"--assembly-policy-workspace", checkpoint, manifest)
	if first.Decision != "PROGRESS" || first.Result.Runtime.FinitePassed != 2 {
		t.Fatal("checkpoint lost partial result")
	}
	path := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	continued, _ := runSavedPackageFixture(t, "resume", "--json", "--receipt", path,
		"--cases", cases, "--assembly-policy-workspace", policy, manifest)
	if continued.Decision != "PASS" || continued.Result.Runtime.FinitePassed != 4 ||
		continued.Result.Composition.Continuation == nil || continued.Result.Composition.Continuation.NewModelCalls != 0 {
		t.Fatal("source-derived workspace or policy did not continue")
	}
}
