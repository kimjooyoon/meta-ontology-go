package textbound

import (
	"encoding/json"
	"errors"
	"testing"
	"unsafe"
)

func TestTextBoundOptionalValueFresh1905(t *testing.T) {
	if unsafe.Sizeof(UTF8(0)) != 8 {
		t.Fatal("bound must remain an eight-byte value")
	}
	for _, tc := range []struct {
		name string
		raw  string
		want uint32
	}{
		{"zero", "0", 0},
		{"ingress", "1024", 1024},
		{"expanded", "1030", 1030},
		{"uint32-maximum", "4294967295", 4294967295},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bound, err := Parse(tc.raw)
			if err != nil || !bound.IsSet() {
				t.Fatal("declared bound lost presence", bound, err)
			}
			actual, err := bound.Resolve(1024)
			if err != nil || actual != tc.want || bound != New(tc.want) {
				t.Fatal("bound changed", actual, err)
			}
			if actual, err := (UTF8(0)).Resolve(1024); err != nil || actual != 1024 {
				t.Fatal("undeclared bound changed legacy limit", actual, err)
			}
		})
	}
}

func TestTextBoundMalformedSourceFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{{"empty", ""}, {"negative", "-1"}, {"plus", "+1"}, {"double-zero", "00"}, {"leading-zero", "01"},
		{"decimal", "1.5"}, {"exponent", "1e3"}, {"leading-space", " 1"}, {"trailing-space", "1 "},
		{"unicode-digit", "１"}, {"uint32-overflow", "4294967296"}} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse(tc.value); !errors.Is(err, ErrInvalid) {
				t.Fatal("invalid source bound accepted", tc.value, err)
			}
		})
	}
	if _, err := (UTF8(1<<32 + 1)).Resolve(1024); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid carrier silently defaulted or wrapped", err)
	}
}

func TestTextBoundJSONPresenceFresh1905(t *testing.T) {
	type record struct {
		Bound UTF8 `json:"max_utf8_bytes,omitempty"`
	}
	for _, tc := range []struct {
		value UTF8
		want  string
	}{{0, `{}`}, {New(0), `{"max_utf8_bytes":0}`}, {New(1030), `{"max_utf8_bytes":1030}`}} {
		raw, err := json.Marshal(record{Bound: tc.value})
		if err != nil || string(raw) != tc.want {
			t.Fatal("serialized storage encoding or erased zero", string(raw), err)
		}
		var replay record
		if err := json.Unmarshal(raw, &replay); err != nil || replay.Bound != tc.value {
			t.Fatal("optional bound changed on round trip", replay, err)
		}
	}
}
