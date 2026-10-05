package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPackageResolveBuildsStableWorkspaceGraph(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Integer id "workspace://core/integer"
activity AddOne(Integer) -> Integer computes "int.add:1"
`)
	writeWorkspaceSource(t, root, "main.gooo", `package app
namespace app
entity Integer id "workspace://app/integer"
activity Clamp(Integer) -> Integer computes "int.add:1"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Clamp"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["main.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	args := []string{"package", "resolve", "--json", manifestPath}
	var firstOut, firstErr, secondOut, secondErr bytes.Buffer
	firstCode := run(args, &firstOut, &firstErr)
	secondCode := run(args, &secondOut, &secondErr)
	if firstCode != exitOK || secondCode != exitOK || firstErr.Len() != 0 || secondErr.Len() != 0 {
		t.Fatalf("resolution failed: first=(%d,%q) second=(%d,%q)", firstCode, firstErr.String(), secondCode, secondErr.String())
	}
	if !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) {
		t.Fatalf("workspace resolution was not deterministic:\nfirst=%s\nsecond=%s", firstOut.String(), secondOut.String())
	}
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(firstOut.Bytes(), &receipt); err != nil {
		t.Fatalf("decode workspace receipt: %v", err)
	}
	if receipt.Decision != "PASS" || receipt.Result == nil || receipt.ManifestDigest == "" {
		t.Fatalf("incomplete workspace receipt: %#v", receipt)
	}
	order := receipt.Result.Image.InitOrder
	if len(order) != 2 || order[0] != "example/core" || order[1] != "example/app" || receipt.Result.Image.Entry.Activity != "Clamp" {
		t.Fatalf("unexpected resolved package graph: order=%v entry=%#v", order, receipt.Result.Image.Entry)
	}
}

func TestRunPackageResolveFailsClosedOnUnknownImport(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "main.gooo", `package app
namespace app
entity Integer id "workspace://app/integer"
activity Clamp(Integer) -> Integer computes "int.add:1"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Clamp"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/missing"], "sources": ["main.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode failure receipt: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, "PACKAGE_IMPORT_UNKNOWN") || stderr.Len() != 0 {
		t.Fatalf("unknown import did not fail closed: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}

func writeWorkspaceSource(t *testing.T, root, name, source string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeWorkspaceManifest(t *testing.T, root, contents string) string {
	t.Helper()
	filename := filepath.Join(root, "gooo.workspace.json")
	if err := os.WriteFile(filename, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return filename
}
