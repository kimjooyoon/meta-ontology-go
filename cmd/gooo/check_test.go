package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestRunCheckValidSource(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"billing.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stdout.String() != "ok: billing.gooo\n" || stderr.Len() != 0 {
		t.Fatalf("check result = code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
}

func TestRunCheckReportsDeterministicDiagnostics(t *testing.T) {
	var firstOut, firstErr bytes.Buffer
	firstCode := runCheck([]string{"broken.gooo"}, fixtureReader{source: "package billing\nentity Broken id \"x\" @"}, SyntaxSourceParser{}, &firstOut, &firstErr)
	var secondOut, secondErr bytes.Buffer
	secondCode := runCheck([]string{"broken.gooo"}, fixtureReader{source: "package billing\nentity Broken id \"x\" @"}, SyntaxSourceParser{}, &secondOut, &secondErr)
	if firstCode != exitFailure || secondCode != exitFailure || firstOut.String() != secondOut.String() || firstErr.String() != secondErr.String() {
		t.Fatalf("diagnostics were not deterministic: first=(%d,%q,%q), second=(%d,%q,%q)", firstCode, firstOut.String(), firstErr.String(), secondCode, secondOut.String(), secondErr.String())
	}
	want := "broken.gooo:2:1-2:7: error parse.expected-namespace: expected namespace declaration\n" +
		"broken.gooo:2:22-2:23: error lex.unexpected-character: unexpected character '@'\n"
	if firstErr.String() != want {
		t.Fatalf("diagnostics = %q, want %q", firstErr.String(), want)
	}
}

func TestRunCheckReadErrorAndUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runCheck([]string{"missing.gooo"}, fixtureReader{err: errors.New("fixture missing")}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || stderr.String() != "gooo: missing.gooo: read error: fixture missing\n" {
		t.Fatalf("read error = code %d, stderr %q", code, stderr.String())
	}
	stderr.Reset()
	code = runCheck(nil, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo check [--semantic] <file.gooo>\n" {
		t.Fatalf("usage = code %d, stderr %q", code, stderr.String())
	}
}

func TestRunCheckSemanticModeIsExplicit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"--semantic", "billing.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stdout.String() != "ok: billing.gooo\n" || stderr.String() != deferredCheckProvenance+"\n" {
		t.Fatalf("semantic check result = code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
}

func TestRunCheckDefaultPreservesSyntaxOnlyMode(t *testing.T) {
	source := `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"billing.gooo"}, fixtureReader{source: source}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stdout.String() != "ok: billing.gooo\n" || stderr.Len() != 0 {
		t.Fatalf("syntax-only check result = code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
}

func TestRunCheckUsesParserSeam(t *testing.T) {
	parser := recordingParser{}
	var stdout, stderr bytes.Buffer
	code := runCheck([]string{"fixture.gooo"}, fixtureReader{source: validSource}, &parser, &stdout, &stderr)
	if code != exitOK || parser.filename != "fixture.gooo" || parser.source != validSource || stdout.String() != "ok: fixture.gooo\n" {
		t.Fatalf("parser seam = code %d, filename %q, source %q, stdout %q", code, parser.filename, parser.source, stdout.String())
	}
}

func TestRunDispatchesCheckAndUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "usage: gooo check [--semantic]") {
		t.Fatalf("check usage = code %d, stderr %q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(nil, &stdout, &stderr); code != exitUsage || stderr.String() != "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|provenance|lsp|version> [args]\n" {
		t.Fatalf("root usage = code %d, stderr %q", code, stderr.String())
	}
}

type fixtureReader struct {
	source string
	err    error
}

func (r fixtureReader) ReadFile(string) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return []byte(r.source), nil
}

type recordingParser struct {
	filename string
	source   string
}

func (p *recordingParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	p.filename = filename
	p.source = source
	return syntax.ParseFile(filename, source)
}

const validSource = `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Order) -> Order
`
