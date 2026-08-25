package main

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
	"testing"
)

func TestRunDispatchesCheckAndUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"check"}, &stdout, &stderr); code != exitUsage || !strings.Contains(stderr.String(), "usage: gooo check [--semantic]") {
		t.Fatalf("check usage = code %d, stderr %q", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(nil, &stdout, &stderr); code != exitUsage || stderr.String() != "usage: gooo <run|profile|debug|emit|check|generate|roundtrip|query|inspect|graph|analyze|format|fix|provenance|selective-ci|lsp|version> [args]\n" {
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
