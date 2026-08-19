package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
	"testing"
)

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
