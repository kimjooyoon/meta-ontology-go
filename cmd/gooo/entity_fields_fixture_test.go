package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type entityFieldsMapReader map[string][]byte

type deferredCLIWorkspace struct {
	parent, outputDir, manifestPath, storePath, tempPath string
}

func (reader entityFieldsMapReader) ReadFile(name string) ([]byte, error) {
	data, ok := reader[name]
	if !ok {
		return nil, fmt.Errorf("missing fixture %q", name)
	}
	return append([]byte(nil), data...), nil
}

func prepareDeferredCLIWorkspace(t *testing.T) deferredCLIWorkspace {
	t.Helper()
	workspace := deferredCLIWorkspace{parent: t.TempDir()}
	workspace.outputDir = filepath.Join(workspace.parent, "output")
	workspace.manifestPath = filepath.Join(workspace.parent, "evidence", "projection.jsonl")
	workspace.storePath = filepath.Join(workspace.parent, "evidence", "provenance.jsonl")
	workspace.tempPath = filepath.Join(workspace.outputDir, ".semantic.gooo.go.tmp-stale")
	for path, data := range map[string][]byte{
		filepath.Join(workspace.outputDir, generatedFileName): []byte("old generated\n"),
		workspace.manifestPath:                                []byte("old manifest\n"),
		workspace.storePath:                                   []byte("old evidence\n"),
		workspace.tempPath:                                    []byte("old temp\n"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	return workspace
}

func assertDeferredCLIWorkspace(t *testing.T, workspace deferredCLIWorkspace) {
	t.Helper()
	for path, want := range map[string][]byte{
		filepath.Join(workspace.outputDir, generatedFileName): []byte("old generated\n"),
		workspace.manifestPath:                                []byte("old manifest\n"),
		workspace.storePath:                                   []byte("old evidence\n"),
		workspace.tempPath:                                    []byte("old temp\n"),
	} {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("deferred CLI changed %s: got=%q err=%v", path, got, err)
		}
	}
}

func cliEntityFieldsFixture(t *testing.T) (semantic.IR, bidir.Model) {
	t.Helper()
	const uri = "entity-fields.gooo"
	entityID := semantic.MustIdentity("billing://entity/order")
	stringID := semantic.MustIdentity("urn:gooo:type:string")
	entitySpan := semantic.Span{File: uri, Start: semantic.Position{Offset: 0, Line: 1, Column: 1}, End: semantic.Position{Offset: 160, Line: 8, Column: 1}}
	semanticField := func(id, name string, start, end int) semantic.Field {
		return semantic.Field{
			ID: semantic.ID(id), Parent: entityID, Name: name, TypeRef: semantic.TypeRef{ID: stringID}, Presence: semantic.Required, Cardinality: semantic.One,
			Span: semantic.Span{File: uri, Start: semantic.Position{Offset: start, Line: 2, Column: start + 1}, End: semantic.Position{Offset: end, Line: 2, Column: end + 1}},
		}
	}
	fields := []semantic.Field{semanticField("billing://field/order-number", "OrderNumber", 10, 60), semanticField("billing://field/customer-name", "CustomerName", 70, 120)}
	ir := semantic.NewIR("billing", "billing")
	if err := ir.AddNode(semantic.Node{ID: entityID, Kind: semantic.Entity, Namespace: "billing", Name: "Order", Fields: fields, Span: entitySpan}); err != nil {
		t.Fatal(err)
	}
	bidirField := func(id, name string, start, end int) bidir.Field {
		span := func(left, right int) bidir.SourceSpan {
			return bidir.SourceSpan{File: uri, Start: left, End: right, StartLine: 2, StartColumn: left + 1, EndLine: 2, EndColumn: right + 1}
		}
		nameStart := start + 6
		nameEnd := nameStart + len(name)
		typeStart := nameEnd + 1
		typeEnd := typeStart + 5
		presenceStart := typeEnd + 1
		presenceEnd := presenceStart + 8
		cardinalityStart := presenceEnd + 1
		cardinalityEnd := cardinalityStart + 4
		return bidir.Field{
			ID: bidir.ID(id), Parent: bidir.ID(entityID), Name: name, TypeRef: semantic.TypeRef{ID: stringID}, TypeRefUse: bidir.TypeRefUse{Form: bidir.TypeRefFormStableID, Spelling: string(stringID), ResolvedID: bidir.ID(stringID), Span: span(typeStart, typeEnd)}, Origin: bidir.FieldOriginSource, Presence: bidir.FieldPresenceRequired, Cardinality: bidir.FieldCardinalityOne,
			Span: span(start, end), IDSpan: span(start, start+5), NameSpan: span(nameStart, nameEnd), TypeRefSpan: span(typeStart, typeEnd), PresenceSpan: span(presenceStart, presenceEnd), CardinalitySpan: span(cardinalityStart, cardinalityEnd),
		}
	}
	model := bidir.Model{Package: "billing", Namespace: "billing", Nodes: []bidir.Node{{ID: bidir.ID(entityID), Kind: bidir.EntityKind, Namespace: "billing", Name: "Order", Fields: []bidir.Field{bidirField("billing://field/order-number", "OrderNumber", 10, 60), bidirField("billing://field/customer-name", "CustomerName", 70, 120)}, Span: bidir.SourceSpan{File: uri, Start: 0, End: 160, StartLine: 1, StartColumn: 1, EndLine: 8, EndColumn: 1}}}}
	return ir, model
}

func filesystemDigest(t *testing.T, root string) string {
	t.Helper()
	var records []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		digest := ""
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			digest = semantic.StableHash(data)
		}
		records = append(records, fmt.Sprintf("%s:%o:%d:%s", filepath.ToSlash(relative), info.Mode().Perm(), info.Size(), digest))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(records, "\n")
}

func TestEntityFieldsDeferredAnalyzeRouteHasNoFixPlanOutput(t *testing.T) {
	var stdout, stderr strings.Builder
	reader := entityFieldsMapReader{"fields.gooo": []byte(deferredEntityFieldsSource)}
	if code := runAnalyze([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred analyze code = %d", code)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "parse.entity-fields-deferred") {
		t.Fatalf("deferred analyze output = stdout %q stderr %q", stdout.String(), stderr.String())
	}
}

func TestGenerateArtifactTransactionRestoresBothOutputsOnLaterRenameFailure(t *testing.T) {
	root := t.TempDir()
	outputPath := filepath.Join(root, generatedFileName)
	manifestPath := filepath.Join(root, generatedManifestFileName)
	oldOutput, oldManifest := []byte("old generated\n"), []byte("old manifest\n")
	if err := os.WriteFile(outputPath, oldOutput, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, oldManifest, 0o640); err != nil {
		t.Fatal(err)
	}
	ops := defaultAtomicFileOps()
	renames := 0
	ops.rename = func(oldPath, newPath string) error {
		renames++
		if renames == 2 {
			return errors.New("injected second rename failure")
		}
		return os.Rename(oldPath, newPath)
	}
	err := writeAtomicFilesWithOps([]atomicWrite{{path: outputPath, data: []byte("new generated\n")}, {path: manifestPath, data: []byte("new manifest\n")}}, ops)
	if err == nil || !strings.Contains(err.Error(), "injected second rename failure") {
		t.Fatalf("transaction error = %v", err)
	}
	gotOutput, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	gotManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotOutput, oldOutput) || !bytes.Equal(gotManifest, oldManifest) {
		t.Fatalf("transaction rollback changed pre-state: output=%q manifest=%q", gotOutput, gotManifest)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("transaction left temporary files: %#v", entries)
	}
}

func TestEntityFieldsCLIStateSwitchIsExhaustive(t *testing.T) {
	current := syntax.CurrentEntityFieldsSupport()
	for _, state := range []syntax.EntityFieldsState{syntax.EntityFieldsDeferred, syntax.EntityFieldsSupported} {
		current.State = state
		if err := validateCLIEntityFieldsSupport(current); err != nil {
			t.Fatalf("known state %q rejected: %v", state, err)
		}
	}
	current.State = "UNKNOWN"
	if err := validateCLIEntityFieldsSupport(current); err == nil || !strings.Contains(err.Error(), "GOOO-EF-V1-UNKNOWN-STATE") {
		t.Fatalf("unknown state was not rejected: %v", err)
	}
}
