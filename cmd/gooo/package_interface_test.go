package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPackageInterfaceIncludesImportedRecordShape(t *testing.T) {
	manifest := filepath.Join("..", "..", "examples", "package-optional-record-flow", "gooo.workspace.json")
	var out, diagnostics bytes.Buffer
	code := run([]string{"package", "interface", "--json", manifest}, &out, &diagnostics)
	if code != exitOK || diagnostics.Len() != 0 {
		t.Fatalf("interface: code=%d stdout=%s stderr=%s", code, out.String(), diagnostics.String())
	}
	var receipt workspaceInterfaceReceipt
	if err := json.Unmarshal(out.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Schema != "gooo/package-workspace-interface-receipt/v1" || receipt.Decision != "PASS" {
		t.Fatal(out.String())
	}
	if receipt.ManifestDigest == "" || receipt.Interface == nil || receipt.Interface.Schema != "gooo/package-interface/v1" ||
		receipt.Interface.Scope != "DECLARED_PACKAGE_INTERFACES" || receipt.Interface.Digest == "" {
		t.Fatal(out.String())
	}
	var found, foundEntry bool
	for _, pkg := range receipt.Interface.Packages {
		for _, decl := range pkg.Declarations {
			if pkg.Path == "example/domain" && decl.Name == "Profile" {
				found = true
				if decl.Shape != "record" || decl.ID != "example://domain/profile" || len(decl.Fields) != 3 {
					t.Fatal(decl)
				}
				for index, field := range decl.Fields {
					wantType := []string{"string", "boolean", "integer"}[index]
					if field.ID == "" || field.TypeID != "urn:gooo:type:"+wantType || field.Presence != "optional" || field.Cardinality != "one" {
						t.Fatal(field)
					}
				}
			}
			if pkg.Path == "example/app" && decl.Name == "Main" {
				foundEntry = true
				if decl.ID == "" || !reflect.DeepEqual(decl.Inputs, []string{"example://domain/profile"}) ||
					decl.Output != "example://domain/profile" {
					t.Fatal(decl)
				}
			}
		}
	}
	if !found || !foundEntry {
		t.Fatal("imported record definition or entry activity was omitted")
	}
}

func TestPackageInterfaceSingleFileAndFailure(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceSource(t, root, "single.gooo", "package demo\nnamespace demo\nentity Text id \"demo://text\"\nactivity Echo(Text, Text) -> Text computes \"return input0\"\n")
	path := writeWorkspaceManifest(t, root, `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"demo","activity":"Echo"},"packages":[{"path":"demo","sources":["single.gooo"]}]}`)
	var out, diagnostics bytes.Buffer
	if code := run([]string{"package", "interface", path}, &out, &diagnostics); code != exitOK ||
		!strings.Contains(out.String(), "Echo") || !strings.Contains(out.String(), "demo://text") {
		t.Fatal(code, out.String(), diagnostics.String())
	}
	if err := os.WriteFile(filepath.Join(root, "single.gooo"), []byte("package demo\nnamespace demo\nactivity Echo(Missing) -> Missing\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := run([]string{"package", "interface", "--json", path}, &out, &diagnostics); code != exitFailure ||
		!strings.Contains(out.String(), `"decision":"FAIL_CLOSED"`) || strings.Contains(out.String(), `"interface":`) {
		t.Fatal(code, out.String(), diagnostics.String())
	}
}
