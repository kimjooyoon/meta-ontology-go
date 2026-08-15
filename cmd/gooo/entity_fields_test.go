package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const deferredEntityFieldsSource = `package billing
namespace billing

entity Order id "billing://entity/order" fields {
    field OrderNumber id "billing://field/order-number" type string required one
}

activity PayOrder(Order) -> Order
`

func TestEntityFieldsDeferredPublicCLIIsExactAndNoWrite(t *testing.T) {
	file, diagnostics := syntax.ParseFile("fields.gooo", deferredEntityFieldsSource)
	if file != nil || len(diagnostics) != 1 || diagnostics[0].Code != syntax.DiagEntityFieldsDeferred {
		t.Fatalf("deferred parser boundary = file %#v diagnostics %#v", file, diagnostics)
	}
	if diagnostics[0].Span.Filename != "fields.gooo" || diagnostics[0].Span.Start.Offset == 0 {
		t.Fatalf("deferred diagnostic is not source-backed: %#v", diagnostics[0])
	}

	workspace := prepareDeferredCLIWorkspace(t)
	reader := entityFieldsMapReader{
		"fields.gooo":   []byte(deferredEntityFieldsSource),
		"evidence.json": []byte(`{"records":[]}`),
	}
	before := filesystemDigest(t, workspace.parent)

	var stdout, stderr bytes.Buffer
	if code := runCheck([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred check code = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "parse.entity-fields-deferred") {
		t.Fatalf("deferred check classification = stdout %q stderr %q", stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{"fields.gooo", "--out", workspace.outputDir, "--manifest", workspace.manifestPath}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred generate code = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "parse.entity-fields-deferred") {
		t.Fatalf("deferred generate classification = stdout %q stderr %q", stdout.String(), stderr.String())
	}
	if code := runRoundTrip([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred roundtrip code = %d", code)
	}
	if code := runQuery([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred query code = %d", code)
	}
	if code := runInspect([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred inspect code = %d", code)
	}
	if code := runGraph([]string{"dump", "fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred graph code = %d", code)
	}
	if code := runAnalyze([]string{"fields.gooo", "generated.go"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred analyze code = %d", code)
	}
	if code := runProvenance([]string{"publish", "fields.gooo", "--store", workspace.storePath, "--evidence", "evidence.json"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred provenance code = %d", code)
	}

	if got := filesystemDigest(t, workspace.parent); got != before {
		t.Fatalf("deferred CLI changed filesystem: before=%s after=%s", before, got)
	}
	assertDeferredCLIWorkspace(t, workspace)
}

func TestEntityFieldsDeferredLSPRoutePublishesOnlySourceDiagnostic(t *testing.T) {
	uri := "file:///fields.gooo"
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 1, "text": deferredEntityFieldsSource},
		}),
		lspRequest(2, "shutdown", nil),
		lspNotification("exit", nil),
	)
	output, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("deferred LSP route = code %d stderr=%q output=%q", code, stderr, output)
	}
	if !bytes.Contains(output, []byte("parse.entity-fields-deferred")) || bytes.Contains(output, []byte(`"result":{"symbols"`)) {
		t.Fatalf("deferred LSP output did not stay diagnostic-only: %s", output)
	}
}

func TestFieldlessBillingProjectionBytesAndEvidenceHashesRemainStable(t *testing.T) {
	root := t.TempDir()
	firstDir := filepath.Join(root, "first")
	secondDir := filepath.Join(root, "second")
	fixture := filepath.Join("..", "..", "examples", "billing", "main.gooo")
	if code := runGenerate([]string{fixture, "--out", firstDir}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("first billing generate = %d", code)
	}
	if code := runGenerate([]string{fixture, "--out", secondDir}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("second billing generate = %d", code)
	}
	firstSource, err := os.ReadFile(filepath.Join(firstDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	secondSource, err := os.ReadFile(filepath.Join(secondDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstSource, secondSource) || semantic.StableHash(firstSource) != "3c0ca7a65301c490a6732d4c8635c0dda5d934bb14a6cf645dddc792fffea5d6" {
		t.Fatalf("fieldless billing source changed: first=%s second=%s", semantic.StableHash(firstSource), semantic.StableHash(secondSource))
	}
	var firstManifest, secondManifest projectionManifest
	firstBytes, err := os.ReadFile(filepath.Join(firstDir, generatedManifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(filepath.Join(secondDir, generatedManifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(firstBytes, &firstManifest); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondBytes, &secondManifest); err != nil {
		t.Fatal(err)
	}
	if firstManifest.SemanticDigest != secondManifest.SemanticDigest || firstManifest.GeneratedDigest != secondManifest.GeneratedDigest || firstManifest.SourceMapDigest != secondManifest.SourceMapDigest || firstManifest.ResponseDigest != secondManifest.ResponseDigest || firstManifest.EvidenceManifest.PayloadSHA256 != secondManifest.EvidenceManifest.PayloadSHA256 {
		t.Fatalf("fieldless billing evidence replay diverged: first=%#v second=%#v", firstManifest, secondManifest)
	}
}

func TestEntityFieldsProjectionPreservesBidirIdentityOrderSpansAndProfile(t *testing.T) {
	ir, sourceModel := cliEntityFieldsFixture(t)
	if _, err := projectionIRFromBidirModel(ir, sourceModel); err == nil || !strings.Contains(err.Error(), "parse.entity-fields-deferred") {
		t.Fatalf("ordinary CLI projection activated fields while deferred: %v", err)
	}
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	projected, err := projectionIRFromBidirModelWithSupport(ir, sourceModel, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCLIEntityFieldsSupport(support); err != nil {
		t.Fatal(err)
	}
	wantProfile := syntax.EntityFieldsProfile{ID: syntax.EntityFieldsProfileID, Version: syntax.EntityFieldsProfileVersion, Digest: syntax.EntityFieldsProfileDigest}
	if support.Profile != wantProfile {
		t.Fatalf("profile = %#v, want %#v", support.Profile, wantProfile)
	}
	if len(projected.Entities) != 1 || len(projected.Entities[0].Fields) != 2 {
		t.Fatalf("projected fields = %#v", projected.Entities)
	}
	for index, got := range projected.Entities[0].Fields {
		want := sourceModel.Nodes[0].Fields[index]
		if got.ID != string(want.ID) || got.Parent != string(want.Parent) || got.Name != want.Name || got.TypeRefID != string(want.TypeRef.ID) || got.Presence != string(want.Presence) || got.Cardinality != string(want.Cardinality) {
			t.Fatalf("field %d lost identity/state: got=%#v want=%#v", index, got, want)
		}
		if got.Source != bidirGeneratorSpan(want.Span) || got.IDSpan != bidirGeneratorSpan(want.IDSpan) || got.NameSpan != bidirGeneratorSpan(want.NameSpan) || got.TypeRefSpan != bidirGeneratorSpan(want.TypeRefSpan) || got.PresenceSpan != bidirGeneratorSpan(want.PresenceSpan) || got.CardinalitySpan != bidirGeneratorSpan(want.CardinalitySpan) || got.NameSource != got.NameSpan {
			t.Fatalf("field %d lost exact source provenance: got=%#v want=%#v", index, got, want)
		}
	}
}

func TestEntityFieldsProjectionRejectsMissingAuthoritativeTypeRefIDWithoutWrites(t *testing.T) {
	ir, sourceModel := cliEntityFieldsFixture(t)
	support := syntax.CurrentEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	sourceModel.Nodes[0].Fields[0].TypeRef.ID = ""
	workspace := prepareDeferredCLIWorkspace(t)
	before := filesystemDigest(t, workspace.parent)
	var stdout, stderr bytes.Buffer
	projected, err := projectionIRFromBidirModelWithSupport(ir, sourceModel, support)
	wantErr := `entity "billing://entity/order" field 0: GOOO-EF-V1-INCOMPLETE-FIELD: field has no authoritative TypeRef.ID`
	if err == nil || err.Error() != wantErr || !reflect.DeepEqual(projected, generator.SemanticIR{}) {
		t.Fatalf("missing authoritative TypeRef.ID result = %#v err=%v, want exact rejection", projected, err)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 || filesystemDigest(t, workspace.parent) != before {
		t.Fatalf("missing authoritative TypeRef.ID changed publication state: stdout=%q stderr=%q before=%s after=%s", stdout.String(), stderr.String(), before, filesystemDigest(t, workspace.parent))
	}
	assertDeferredCLIWorkspace(t, workspace)
}

func TestEntityFieldsProjectionRejectsMalformedPartitionsAndUnknownState(t *testing.T) {
	baseIR, baseModel := cliEntityFieldsFixture(t)
	supported := syntax.CurrentEntityFieldsSupport()
	supported.State = syntax.EntityFieldsSupported
	cases := []struct {
		name string
		edit func(*semantic.IR, *bidir.Model, *syntax.EntityFieldsSupport)
		want string
	}{
		{name: "missing ID", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].ID = ""
		}, want: "GOOO-EF-V1"},
		{name: "duplicate ID", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[1].ID = model.Nodes[0].Fields[0].ID
		}, want: "duplicate"},
		{name: "wrong parent", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].Parent = "billing://entity/other"
		}, want: "parent"},
		{name: "wrong snapshot", edit: func(_ *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].IDSpan.File = "other.gooo"
		}, want: "source"},
		{name: "unsupported shape", edit: func(ir *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].Presence = bidir.FieldPresenceOptional
			node, _ := ir.Graph.Node("billing://entity/order")
			node.Fields[0].Presence = semantic.Optional
			ir.Graph = semantic.NewGraph()
			_ = ir.AddNode(node)
		}, want: "UNSUPPORTED-SHAPE"},
		{name: "unsupported type", edit: func(ir *semantic.IR, model *bidir.Model, _ *syntax.EntityFieldsSupport) {
			model.Nodes[0].Fields[0].TypeRef.ID = "urn:gooo:type:integer"
			model.Nodes[0].Fields[0].TypeRefUse = bidir.TypeRefUse{Form: bidir.TypeRefFormStableID, Spelling: "urn:gooo:type:integer", ResolvedID: "urn:gooo:type:integer", Span: model.Nodes[0].Fields[0].TypeRefSpan}
			node, _ := ir.Graph.Node("billing://entity/order")
			node.Fields[0].TypeRef.ID = "urn:gooo:type:integer"
			ir.Graph = semantic.NewGraph()
			_ = ir.AddNode(node)
		}, want: "UNKNOWN-TYPE"},
		{name: "profile mismatch", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) {
			support.Profile.Digest = "tampered"
		}, want: "PROFILE-MISMATCH"},
		{name: "unknown state", edit: func(_ *semantic.IR, _ *bidir.Model, support *syntax.EntityFieldsSupport) { support.State = "UNKNOWN" }, want: "UNKNOWN-STATE"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ir, model := baseIR, baseModel
			normalized, err := ir.Normalized()
			if err != nil {
				t.Fatal(err)
			}
			ir = normalized
			model = model.Clone()
			support := supported
			testCase.edit(&ir, &model, &support)
			projected, err := projectionIRFromBidirModelWithSupport(ir, model, support)
			if err == nil || !strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper(testCase.want)) || !reflect.DeepEqual(projected, generator.SemanticIR{}) {
				t.Fatalf("partition result = %#v err=%v, want error containing %q and empty projection", projected, err, testCase.want)
			}
		})
	}
}
