package main

import (
	"strings"
	"testing"
)

func TestFailureManifestRejectsMismatchedJobBinding(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.Job.HeadSHA = strings.Repeat("b", 40)
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("mismatched failure job head was accepted")
	}
}
func TestFailureManifestRejectsUnknownBinding(t *testing.T) {
	binding := validFailureBinding()
	binding.Actor = "unknown"
	if _, err := buildFailureManifest(validFailureInput(), binding); err == nil {
		t.Fatal("unknown agent binding was accepted")
	}
}
