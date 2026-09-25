package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicGraphRejectsInvalidRelayFields(t *testing.T) {
	source, err := os.ReadFile(relayGraphFixture)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutation := range map[string][2]string{
		"unknown-type":    {"type string", "type missing_type"},
		"duplicate-id":    {"request/run-id", "request/turn"},
		"bad-cardinality": {"required one", "required invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "invalid.gooo")
			mutated := strings.Replace(string(source), mutation[0], mutation[1], 1)
			if mutated == string(source) {
				t.Fatal("counterexample mutation did not change the fixture")
			}
			if err := os.WriteFile(path, []byte(mutated), 0600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := run([]string{"graph", "dump", path}, &stdout, &stderr); code != exitFailure || stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("invalid graph escaped: exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
			}
		})
	}
}
