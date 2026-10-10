package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBodyOutcomesDeltaCLI(t *testing.T) {
	dir := t.TempDir()
	before, after := filepath.Join(dir, "before.json"), filepath.Join(dir, "after.json")
	fixture := `{"schema":"gooo/body-composition-runtime/v1","stage":"COMPLETE","finite_passed":1,"finite_total":1,"traces":[{"case_index":0,"deliveries":[{"activity_id":"test://main","input":9007199254740993,"actual":16,"expected":16,"passed":true}]}]}`
	for path, value := range map[string]string{before: fixture, after: strings.ReplaceAll(fixture, `:16`, `:32`)} {
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out, stderr bytes.Buffer
	args := []string{"--before", before, "--after", after}
	if code := runExtensionCommand(append([]string{"body-outcomes-delta"}, args...), &out, &stderr); code != exitOK || stderr.Len() != 0 || !strings.Contains(out.String(), "Requirements changed: 1; regressions under the same requirement: 0") {
		t.Fatalf("%d %s %s", code, &out, &stderr)
	}
	out.Reset()
	stderr.Reset()
	if code := runBodyOutcomesDelta(append(args, "--json"), &out, &stderr); code != exitOK || !strings.Contains(out.String(), `"assessment":"REQUIREMENT_CHANGED"`) || !strings.Contains(out.String(), "9007199254740993") {
		t.Fatalf("%d %s %s", code, &out, &stderr)
	}
	for _, args := range [][]string{nil, {"--json", "--json"}, {"--before", before}, {"--before", before, "--before", before}, {"--after"}, {"--unknown", "x"}} {
		out.Reset()
		stderr.Reset()
		if code := runBodyOutcomesDelta(args, &out, &stderr); code != exitUsage || out.Len() != 0 {
			t.Fatal("accepted invalid flags")
		}
	}
	out.Reset()
	stderr.Reset()
	if code := runBodyOutcomesDelta([]string{"--before", dir, "--after", after}, &out, &stderr); code != exitFailure || out.Len() != 0 {
		t.Fatal("accepted a directory")
	}
	out.Reset()
	stderr.Reset()
	if code := runHelp([]string{"body-outcomes-delta"}, &out, &stderr); code != exitOK || !strings.Contains(out.String(), "complete caller input tuples") {
		t.Fatal("missing help")
	}
}
