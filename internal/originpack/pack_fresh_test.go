package originpack

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRawOriginPackFresh1805(t *testing.T) {
	for name, raw := range map[string][]byte{
		"empty": {}, "bytes": {0, 255, 13, 10, 0},
		"unicode":  []byte("출처\r\norigin\n  \t<>&\u2028"),
		"repeated": bytes.Repeat([]byte("gooo://origin-lens/source 한글\r\n"), 1024),
		"at-bound": bytes.Repeat([]byte{31}, MaxRawBytes),
	} {
		t.Run(name, func(t *testing.T) {
			packed, err := Pack(raw)
			if err != nil {
				t.Fatal(err)
			}
			restored, err := Unpack(packed)
			if err != nil || !bytes.Equal(raw, restored) {
				t.Fatalf("raw bytes not restored: %v", err)
			}
			if name == "repeated" && len(packed) >= len(raw) {
				t.Fatal("repeated corpus did not compress")
			}
		})
	}
}

func TestRawOriginEnvelopeRefusalsFresh1806(t *testing.T) {
	packed, err := Pack(bytes.Repeat([]byte("provenance"), 100))
	if err != nil {
		t.Fatal(err)
	}
	corrupt := bytes.Clone(packed)
	corrupt[13] ^= 1
	oversize := bytes.Clone(packed)
	binary.LittleEndian.PutUint32(oversize[9:13], MaxRawBytes+1)
	for name, input := range map[string][]byte{
		"truncated": packed[:12], "raw-digest": corrupt, "expanded-bound": oversize,
		"trailing-data": append(bytes.Clone(packed), 1), "missing-footer": packed[:len(packed)-4],
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Unpack(input); err == nil {
				t.Fatal("invalid raw origin accepted")
			}
		})
	}
	if _, err := Pack(make([]byte, MaxRawBytes+1)); err == nil {
		t.Fatal("oversize raw origin accepted")
	}
}
