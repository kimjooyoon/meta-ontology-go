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
	actual, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(filepath.Join(directory, "registry.generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatal("registry projection differs from its executable Gooo declaration")
	}
	if err := protectedregions.Check(actual); err != nil {
		t.Fatal(err)
	}
	if bytes.Count(actual, []byte("//gooo:generated:start")) != 6 {
		t.Fatal("expected four activity and two record regions")
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
