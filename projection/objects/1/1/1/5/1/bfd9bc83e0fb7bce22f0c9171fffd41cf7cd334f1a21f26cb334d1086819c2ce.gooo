package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRunInspectG1ContractAndRepeatDeterminism(t *testing.T) {
	firstOutput := inspectFixtureOutput(t, sourceOrderA)
	repeatOutput := inspectFixtureOutput(t, sourceOrderA)
	if !bytes.Equal(firstOutput, repeatOutput) {
		t.Fatalf("repeated inspect output differs:\nfirst=%s\nrepeat=%s", firstOutput, repeatOutput)
	}
	dump := decodeGraphDump(t, firstOutput)
	if dump.SchemaVersion != graphDumpSchemaVersion || dump.GraphHash == "" || dump.SourceDigest == "" {
		t.Fatalf("incomplete graph identity: %#v", dump)
	}
	if dump.IR.Status != "available" || dump.IR.SemanticDigest == "" {
		t.Fatalf("unexpected IR status: %#v", dump.IR)
	}
	if dump.Evidence.Status != "missing" || dump.Evidence.Reason == "" {
		t.Fatalf("unexpected evidence status: %#v", dump.Evidence)
	}
	if dump.Provenance.Status != "missing" || dump.Provenance.Reason == "" {
		t.Fatalf("unexpected provenance status: %#v", dump.Provenance)
	}
	if dump.Projection.Status != "deferred" || dump.Projection.Reason == "" {
		t.Fatalf("unexpected projection status: %#v", dump.Projection)
	}
	if dump.Lowering.Status != "deferred" || dump.Lowering.Reason == "" || dump.Output.Status != "deferred" || dump.Output.Reason == "" {
		t.Fatalf("unexpected lifecycle status: lowering=%#v output=%#v", dump.Lowering, dump.Output)
	}
	wantAuthorities := graphAuthorities{
		GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
		Provenance: "authoritative", Graph: "derived",
	}
	if dump.Authorities != wantAuthorities {
		t.Fatalf("authorities = %#v, want %#v", dump.Authorities, wantAuthorities)
	}
}
func TestRunInspectReportsDiagnosticsAndFailureExit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runInspect([]string{"broken.gooo"}, fixtureReader{
		source: "package billing\nentity Broken id \"x\" @",
	}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitFailure || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("inspect diagnostics = code %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}
func TestRunInspectReadErrorAndUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runInspect([]string{"missing.gooo"}, fixtureReader{err: os.ErrNotExist}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || stderr.Len() == 0 {
		t.Fatalf("inspect read error = code %d, stderr=%q", code, stderr.String())
	}
	stderr.Reset()
	code = runInspect(nil, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo inspect <file.gooo>\n" {
		t.Fatalf("inspect usage = code %d, stderr=%q", code, stderr.String())
	}
}
