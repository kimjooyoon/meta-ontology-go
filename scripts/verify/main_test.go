package main

import (
	"strings"
	"testing"
)

func TestRevisionAvailabilityFailsClosed(t *testing.T) {
	if revisionAvailable(".", "") {
		t.Fatal("empty revision was accepted")
	}
	if revisionAvailable(".", strings.Repeat("0", 40)) {
		t.Fatal("zero revision was accepted")
	}
	if revisionAvailable(".", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef") {
		t.Fatal("missing revision was accepted")
	}
}
