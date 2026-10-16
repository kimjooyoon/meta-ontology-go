package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/protectedregions"
)

func TestRegistryProjectionFresh1631(t *testing.T) {
	directory := "../../internal/meta/metricprogram"
	output := filepath.Join(t.TempDir(), "registry.generated.go")
	if err := generate(filepath.Join(directory, "registry.gooo"), output); err != nil {
		t.Fatal(err)
	}
	actual, expected := projectionFiles(t, filepath.Dir(output)), projectionFiles(t, directory)
	if len(actual) != len(expected) {
		t.Fatal("registry projection file population differs")
	}
	regions := 0
	for name, content := range actual {
		if !bytes.Equal(content, expected[name]) {
			t.Fatal("registry projection differs from its executable Gooo declaration:", name)
		}
		if filepath.Ext(name) == ".go" {
			if err := protectedregions.Check(content); err != nil {
				t.Fatal(err)
			}
			regions += bytes.Count(content, []byte("//gooo:generated:start"))
		}
	}
	if regions != 9 {
		t.Fatal("expected six activity and three record regions")
	}
}

func TestRegistryTargetFieldsFresh1631(t *testing.T) {
	for _, item := range []struct{ name, source string }{
		{"absent", "package fixture; type Other struct { Value string }"},
		{"duplicate", "package fixture; type Row struct {}; type Row struct {}"},
		{"grouped", "package fixture; type Row struct { A, B string }"},
		{"nested", "package fixture; type Row struct { Value []string }"},
	} {
		t.Run(item.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "row.go")
			if err := os.WriteFile(path, []byte(item.source), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := targetFields(path, "Row"); err == nil {
				t.Fatal("ambiguous or non-scalar target accepted")
			}
		})
	}
}
