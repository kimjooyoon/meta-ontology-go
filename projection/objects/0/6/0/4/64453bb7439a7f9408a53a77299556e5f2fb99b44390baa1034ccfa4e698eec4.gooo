package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestRequestCanonicalizationIsOrderIndependent(t *testing.T) {
	left := sampleRequest(StatusPass)
	right := sampleRequest(StatusPass)
	left.Input.IR = []byte(`{"z":2,"a":{"y":4,"x":3}}`)
	right.Input.IR = []byte(`{"a":{"x":3,"y":4},"z":2}`)
	leftPayload, err := left.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	rightPayload, err := right.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftPayload, rightPayload) {
		t.Fatalf("equivalent IR was not canonicalized:\n%s\n%s", leftPayload, rightPayload)
	}
	leftDigest, err := left.Digest()
	if err != nil {
		t.Fatal(err)
	}
	rightDigest, err := right.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if leftDigest != rightDigest || len(leftDigest) != 64 {
		t.Fatalf("canonical digest mismatch: %q != %q", leftDigest, rightDigest)
	}
	left.Input.IR = []byte(`{"a":1}{"b":2}`)
	if _, err := left.CanonicalJSON(); err == nil {
		t.Fatal("trailing JSON was accepted")
	}
}
func TestResponseCanonicalizationSortsProtocolSets(t *testing.T) {
	left := sampleResponse(StatusPass, false)
	right := sampleResponse(StatusPass, true)
	leftPayload, err := left.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	rightPayload, err := right.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(leftPayload, rightPayload) {
		t.Fatalf("equivalent response was not canonicalized:\n%s\n%s", leftPayload, rightPayload)
	}
	if got := string(leftPayload); !strings.HasSuffix(got, "\n") {
		t.Fatalf("canonical response is not JSONL: %q", got)
	}
}
