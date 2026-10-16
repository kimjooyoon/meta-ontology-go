package main

import (
	"bytes"
	"slices"
	"testing"
)

func TestCanonicalTransportPathsFresh1697(t *testing.T) {
	for _, tc := range []struct {
		name, replacement string
	}{
		{"parent", "../program.gooo"},
		{"absolute", "/program.gooo"},
		{"backslash", `..\program.gooo`},
		{"normalized-parent", "safe/../program.gooo"},
		{"duplicate", "program.gooo"},
		{"unknown", "other.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			names := transportNames()
			names[len(names)-1] = tc.replacement
			cfg := transportFixtureFresh1677(t, names)
			if _, err := transportPayloads(cfg); err == nil || err.Error() != "producer archive has an unbound member" {
				t.Fatalf("unbound path must fail before payload reads: %v", err)
			}
		})
	}
}

func TestCanonicalTransportOrderFresh1697(t *testing.T) {
	names := transportNames()
	slices.Reverse(names)
	files, err := transportPayloads(transportFixtureFresh1677(t, names))
	if err != nil || len(files) != len(transportNames()) {
		t.Fatalf("canonical member set rejected: %v", err)
	}
	for _, name := range transportNames() {
		if !bytes.Equal(files[name], []byte("transport:"+name)) {
			t.Fatalf("canonical path lost its archive bytes: %s", name)
		}
	}
}
