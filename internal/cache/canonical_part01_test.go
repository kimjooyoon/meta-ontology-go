package cache

import (
	"bytes"
	"math"
	"testing"
	"time"
)

func TestCanonicalMapOrderingAndTypeTags(t *testing.T) {
	first := map[string]any{"b": int(2), "a": []string{"x", "y"}}
	second := map[string]any{"a": []string{"x", "y"}, "b": int(2)}
	firstBytes, err := CanonicalBytes(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := CanonicalBytes(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("map insertion order changed canonical bytes")
	}
	intDigest, err := DigestOf(int(1))
	if err != nil {
		t.Fatal(err)
	}
	floatDigest, err := DigestOf(float64(1))
	if err != nil {
		t.Fatal(err)
	}
	if intDigest == floatDigest {
		t.Fatal("different scalar types shared a digest")
	}
}
func TestCanonicalPreservesNilFloatAndTimeSemantics(t *testing.T) {
	var nilSlice []byte
	emptySlice := []byte{}
	nilBytes, err := CanonicalBytes(nilSlice)
	if err != nil {
		t.Fatal(err)
	}
	emptyBytes, err := CanonicalBytes(emptySlice)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(nilBytes, emptyBytes) {
		t.Fatal("nil and empty slices shared canonical bytes")
	}
	negativeZero, err := DigestOf(math.Copysign(0, -1))
	if err != nil {
		t.Fatal(err)
	}
	positiveZero, err := DigestOf(float64(0))
	if err != nil {
		t.Fatal(err)
	}
	if negativeZero == positiveZero {
		t.Fatal("signed zero lost its canonical distinction")
	}
	instant := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	if digest, err := DigestOf(instant); err != nil || !digest.Valid() {
		t.Fatalf("canonical time digest invalid: %v %q", err, digest)
	}
}
