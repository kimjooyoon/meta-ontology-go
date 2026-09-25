package main

import (
	"bytes"
	"testing"
)

func TestRunGraphDumpIsCanonicalInspectAlias(t *testing.T) {
	var inspectOut, inspectErr bytes.Buffer
	inspectCode := runInspect([]string{"fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &inspectOut, &inspectErr)
	var graphOut, graphErr bytes.Buffer
	graphCode := runGraph([]string{"dump", "fixture.gooo"}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &graphOut, &graphErr)
	if inspectCode != exitOK || graphCode != exitOK || inspectErr.Len() != 0 || graphErr.Len() != 0 {
		t.Fatalf("graph dump codes = inspect %d/graph %d, errors=%q/%q", inspectCode, graphCode, inspectErr.String(), graphErr.String())
	}
	if !bytes.Equal(inspectOut.Bytes(), graphOut.Bytes()) {
		t.Fatalf("graph dump differs from canonical inspect output:\ninspect=%s\ngraph=%s", inspectOut.Bytes(), graphOut.Bytes())
	}
}

func TestRunGraphDumpUsage(t *testing.T) {
	var stderr bytes.Buffer
	code := runGraph([]string{"query", "fixture.gooo"}, fixtureReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitUsage || stderr.String() != "usage: gooo graph dump <file.gooo>\n" {
		t.Fatalf("graph usage = code %d, stderr=%q", code, stderr.String())
	}
}
