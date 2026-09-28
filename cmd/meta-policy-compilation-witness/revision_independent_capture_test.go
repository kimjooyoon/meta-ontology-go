package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func TestRevisionConsumerSnapshotDoesNotFollowCallerReplacement(t *testing.T) {
	original := filepath.Join(t.TempDir(), "consumer")
	data := []byte("explicit executable snapshot fixture, never executed")
	if err := os.WriteFile(original, data, 0600); err != nil {
		t.Fatal(err)
	}
	snapshot, digest, err := pinRevisionConsumer(t.TempDir(), original, policycompilation.DigestBytes(data))
	if err != nil || digest != policycompilation.DigestBytes(data) {
		t.Fatal("could not pin consumer fixture")
	}
	if err := os.WriteFile(original, []byte("replaced"), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := os.ReadFile(snapshot)
	if err != nil || !bytes.Equal(captured, data) {
		t.Fatal("caller replacement changed the executable snapshot")
	}
}

func TestRevisionConsumerOutputCaptureIsBounded(t *testing.T) {
	capture := &revisionBoundedCapture{limit: 3}
	if n, err := capture.Write([]byte("abcd")); n != 3 || err == nil || capture.String() != "abc" {
		t.Fatal("over-limit consumer output was accepted or discarded without a bound")
	}
}
