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
import "example/core"
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

func TestWorkspaceSourcePathAcceptsGoooFixtures(t *testing.T) {
	got, err := workspaceSourcePath("examples/package-imports/app.gooo.fixture")
	if err != nil || got != "examples/package-imports/app.gooo.fixture" {
		t.Fatalf("workspace fixture path was rejected or changed: got=%q err=%v", got, err)
	}
	if _, err := workspaceSourcePath("examples/package-imports/app.txt"); err == nil {
		t.Fatal("workspace accepted a source with an unsupported extension")
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

func TestRunPackageResolvePublishesCrossPackageActivityTypes(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Money id "workspace://core/money"
activity Normalize(Money) -> Money computes "identity"
`)
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import "example/core"
entity Receipt id "workspace://app/receipt"
activity Charge(Money) -> Receipt computes "identity"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Charge"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["app.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("workspace type resolution failed: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode workspace receipt: %v", err)
	}
	if receipt.Result == nil {
		t.Fatalf("workspace receipt has no result: %+v", receipt)
	}
	for _, pkg := range receipt.Result.Image.Packages {
		if pkg.Path != "example/app" {
			continue
		}
		for _, export := range pkg.Exports {
			if export.Name == "Charge" && export.Kind == "activity" && len(export.InputTypes) == 1 && export.InputTypes[0] == "workspace://core/money" && export.OutputType == "workspace://app/receipt" {
				return
			}
		}
		t.Fatalf("app package did not expose resolved activity types: %+v", pkg.Exports)
	}
	t.Fatalf("workspace result did not include app package: %+v", receipt.Result.Image.Packages)
}

func TestRunPackageResolveChecksImportedActivityBinding(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Text id "workspace://core/text"
activity Normalize(Text) -> Text computes "identity"
`)
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import core "example/core"
activity Main(Text) -> Text computes "identity"
bind core.Normalize.result -> Main.input
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Main"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["app.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("imported activity binding did not resolve: code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
	}
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	if receipt.Result == nil {
		t.Fatalf("workspace receipt has no result: %+v", receipt)
	}
	for _, pkg := range receipt.Result.Image.Packages {
		if pkg.Path != "example/app" {
			continue
		}
		if len(pkg.Bindings) != 1 {
			t.Fatalf("expected one cross-package edge, got %+v", pkg.Bindings)
		}
		binding := pkg.Bindings[0]
		if binding.ProducerPackage != "example/core" || binding.ProducerActivity != "Normalize" || binding.ProducerPort != "result" || binding.ConsumerPackage != "example/app" || binding.ConsumerActivity != "Main" || binding.ConsumerPort != "input" || binding.EntityID != "workspace://core/text" {
			t.Fatalf("unexpected cross-package edge: %+v", binding)
		}
		return
	}
	t.Fatalf("workspace result omitted app package: %+v", receipt.Result.Image.Packages)
}

func TestRunPackageResolveExplainsExportsAndImportedBindings(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Text id "workspace://core/text"
activity Normalize(Text) -> Text computes "identity"
`)
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import core "example/core"
activity Main(Text) -> Text computes "identity"
bind core.Normalize.result -> Main.input
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Main"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["app.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "resolve", manifestPath}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("workspace resolution failed: code=%d stderr=%q", code, stderr.String())
	}
	for _, want := range []string{
		"package example/core (core)",
		"entity Text id=workspace://core/text",
		"activity Normalize(workspace://core/text) -> workspace://core/text",
		"package example/app (app)",
		"binding example/core.Normalize.result -> example/app.Main.input (workspace://core/text)",
		"does not execute activity bodies",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("package explanation omitted %q:\n%s", want, stdout.String())
		}
	}
}

func TestRunPackageResolveRejectsMismatchedImportedActivityBinding(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Text id "workspace://core/text"
activity Normalize(Text) -> Text computes "identity"
`)
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import core "example/core"
entity Boolean id "workspace://app/boolean"
activity Main(Boolean) -> Boolean computes "identity"
bind core.Normalize.result -> Main.input
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Main"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["app.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode failure receipt: %v stdout=%q", err, stdout.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, "PACKAGE_BINDING_TYPE_MISMATCH") || stderr.Len() != 0 {
		t.Fatalf("mismatched cross-package edge did not fail closed: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}

func TestRunPackageResolveFailsClosedOnUnknownActivityType(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "main.gooo", `package app
namespace app
activity Run(Missing) -> Missing computes "identity"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Run"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": [], "sources": ["main.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode failure receipt: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, "PACKAGE_TYPE_UNKNOWN") || stderr.Len() != 0 {
		t.Fatalf("unknown activity type did not fail closed: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}

func TestRunPackageResolveFailsClosedOnAmbiguousImportedType(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core-a.gooo", `package corea
namespace corea
entity Money id "workspace://core-a/money"
`)
	writeWorkspaceSource(t, root, "core-b.gooo", `package coreb
namespace coreb
entity Money id "workspace://core-b/money"
`)
	writeWorkspaceSource(t, root, "app.gooo", `package app
namespace app
import "example/core-a"
import "example/core-b"
activity Run(Money) -> Money computes "identity"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Run"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core-a", "example/core-b"], "sources": ["app.gooo"]},
    {"path": "example/core-a", "name": "corea", "imports": [], "sources": ["core-a.gooo"]},
    {"path": "example/core-b", "name": "coreb", "imports": [], "sources": ["core-b.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode failure receipt: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, "PACKAGE_TYPE_AMBIGUOUS") || stderr.Len() != 0 {
		t.Fatalf("ambiguous imported type did not fail closed: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}

func TestRunPackageResolveFailsClosedWhenSourceImportsDifferFromManifest(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "main.gooo", `package app
namespace app
import "example/core"
entity Integer id "workspace://app/integer"
activity Run(Integer) -> Integer computes "identity"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Run"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": [], "sources": ["main.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Integer id "workspace://core/integer"
`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode failure receipt: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if code != exitFailure || receipt.Decision != "FAIL_CLOSED" || !strings.Contains(receipt.Error, "PACKAGE_SOURCE_IMPORT_MISMATCH") || stderr.Len() != 0 {
		t.Fatalf("source/manifest import mismatch did not fail closed: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
	}
}

func TestRunPackageResolveKeepsManifestOnlyImportCompatibility(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "core.gooo", `package core
namespace core
entity Integer id "workspace://core/integer"
`)
	writeWorkspaceSource(t, root, "main.gooo", `package app
namespace app
entity Integer id "workspace://app/integer"
activity Run(Integer) -> Integer computes "identity"
`)
	manifestPath := writeWorkspaceManifest(t, root, `{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "example/app", "activity": "Run"},
  "packages": [
    {"path": "example/app", "name": "app", "imports": ["example/core"], "sources": ["main.gooo"]},
    {"path": "example/core", "name": "core", "imports": [], "sources": ["core.gooo"]}
  ]
}`)
	var stdout, stderr bytes.Buffer
	code := run([]string{"package", "resolve", "--json", manifestPath}, &stdout, &stderr)
	var receipt workspaceResolutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatalf("decode workspace receipt: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if code != exitOK || receipt.Decision != "PASS" || receipt.Result == nil || stderr.Len() != 0 {
		t.Fatalf("manifest-only imports lost compatibility: code=%d receipt=%+v stderr=%q", code, receipt, stderr.String())
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
