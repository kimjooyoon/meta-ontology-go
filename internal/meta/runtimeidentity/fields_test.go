package runtimeidentity

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
)

func TestCommittedDescriptionsAreProjectedFromGooo(t *testing.T) {
	source, err := os.ReadFile("fields.gooo")
	if err != nil {
		t.Fatal(err)
	}
	want, err := bodylibrary.Generate(context.Background(), "internal/meta/runtimeidentity/fields.gooo", source,
		"urn:gooo:runtime-identity:describe", "runtimeidentity")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("semantic.gooo.go")
	if err != nil || !bytes.Equal(got, []byte(want)) {
		t.Fatal("runtime field descriptions differ from their source projection", err)
	}
}
