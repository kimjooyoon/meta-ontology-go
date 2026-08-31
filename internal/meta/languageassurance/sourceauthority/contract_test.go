package sourceauthority

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmbeddedContractIsValid(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}
	if contract.MetricID != MetricID {
		t.Fatalf("metric id = %q", contract.MetricID)
	}
	if got, want := Digest(), digestBytes(Snapshot()); got != want {
		t.Fatalf("digest = %q want %q", got, want)
	}
	if !strings.HasPrefix(Digest(), "sha256:") {
		t.Fatalf("digest is not content addressed: %q", Digest())
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	raw := bytes.Replace(
		Snapshot(),
		[]byte("{"),
		[]byte(`{"unexpected":true,`),
		1,
	)
	if _, err := Decode(raw); err == nil {
		t.Fatal("unknown field was accepted")
	}
}
