package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountercaseTransportPopulationFresh1679(t *testing.T) {
	for _, mode := range []string{"exact", "changed-unpacked", "missing-unpacked", "changed-archive"} {
		t.Run(mode, func(t *testing.T) {
			cfg := transportFixtureFresh1677(t, transportNames())
			switch mode {
			case "changed-unpacked":
				writeFixtureFresh1677(t, filepath.Join(cfg.Input, "program.gooo"), []byte("changed"))
			case "missing-unpacked":
				if err := os.Remove(filepath.Join(cfg.Input, "program.gooo")); err != nil {
					t.Fatal(err)
				}
			case "changed-archive":
				writeFixtureFresh1677(t, cfg.Archive, []byte("changed"))
			}
			_, err := transportPayloads(cfg)
			if (err == nil) != (mode == "exact") {
				t.Fatalf("archive/unpacked identity: %v", err)
			}
		})
	}
}

func TestCountercaseArchiveRefusalsFresh1679(t *testing.T) {
	for _, mode := range []string{"extra", "duplicate", "traversal"} {
		t.Run(mode, func(t *testing.T) {
			names := transportNames()
			switch mode {
			case "extra":
				names = append(names, "unlisted.json")
			case "duplicate":
				names[len(names)-1] = names[0]
			case "traversal":
				names[len(names)-1] = "../program.gooo"
			}
			if _, err := transportPayloads(transportFixtureFresh1677(t, names)); err == nil {
				t.Fatal("unbound archive member set was accepted")
			}
		})
	}
}
