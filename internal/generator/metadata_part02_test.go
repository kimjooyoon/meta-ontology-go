package generator

import (
	"bytes"
	"reflect"
	"testing"
)

func TestGenerateWithMetadataRejectsInvalidIRWithoutMetadata(t *testing.T) {
	ir := acceptanceFixture()
	ir.Activities[0].Slots[0].ID = ""
	if result, err := GenerateWithMetadata(ir, nil); err == nil || result.Metadata.SourceDigest != "" {
		t.Fatalf("invalid IR returned metadata: result=%#v err=%v", result, err)
	}
}
func TestGenerateWithMetadataDoesNotMutateIROrPrevious(t *testing.T) {
	ir := acceptanceFixture()
	beforeIR := copyIR(ir)
	initial, err := Generate(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	previous := append([]byte(nil), initial.Source...)
	beforePrevious := append([]byte(nil), previous...)
	if _, err := GenerateWithMetadata(ir, previous); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ir, beforeIR) {
		t.Fatal("metadata generation mutated caller-owned IR")
	}
	if !bytes.Equal(previous, beforePrevious) {
		t.Fatal("metadata generation mutated caller-owned previous source")
	}
}
