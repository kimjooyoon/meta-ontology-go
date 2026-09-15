package cache

import (
	"bytes"
	"testing"
)

func TestCanonicalRejectsUnsupportedAndCyclicValues(t *testing.T) {
	if _, err := CanonicalBytes(func() {}); err == nil {
		t.Fatal("function value was accepted")
	}
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	if _, err := CanonicalBytes(cyclic); err == nil {
		t.Fatal("cyclic map was accepted")
	}
	type hidden struct{ value int }
	if _, err := CanonicalBytes(hidden{value: 1}); err == nil {
		t.Fatal("unexported field was accepted")
	}
}
func TestDigestValidityAndCopy(t *testing.T) {
	digest := HashBytes([]byte("hello"))
	if !digest.Valid() || digest.String() != string(digest) {
		t.Fatalf("invalid digest representation: %q", digest)
	}
	if Digest("ABC").Valid() || Digest("not-a-digest").Valid() {
		t.Fatal("invalid digest was accepted")
	}
	bytesValue, err := CanonicalBytes("stable")
	if err != nil {
		t.Fatal(err)
	}
	bytesValue[0] ^= 1
	again, err := CanonicalBytes("stable")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(bytesValue, again) {
		t.Fatal("canonical output unexpectedly aliases encoder state")
	}
}
