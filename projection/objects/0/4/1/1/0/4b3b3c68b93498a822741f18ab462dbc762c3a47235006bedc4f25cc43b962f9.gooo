package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
)

// Digest is a lowercase hexadecimal SHA-256 digest.
type Digest string

const digestLength = sha256.Size * 2

// String returns the hexadecimal representation of the digest.
func (d Digest) String() string { return string(d) }

// Valid reports whether d is a complete lowercase hexadecimal SHA-256 digest.
func (d Digest) Valid() bool {
	if len(d) != digestLength || strings.ToLower(string(d)) != string(d) {
		return false
	}
	_, err := hex.DecodeString(string(d))
	return err == nil
}

// Known reports whether d is valid and not the zero sentinel used for
// missing evidence.
func (d Digest) Known() bool {
	return d.Valid() && d != Digest(strings.Repeat("0", digestLength))
}

// HashBytes returns the SHA-256 digest of data. The result is safe to use as a
// content-addressed filename.
func HashBytes(data []byte) Digest {
	sum := sha256.Sum256(data)
	return Digest(hex.EncodeToString(sum[:]))
}

// CanonicalBytes encodes a value deterministically for hashing. Maps and
// struct fields are sorted, while scalar types retain their Go type.
func CanonicalBytes(value any) ([]byte, error) {
	encoder := canonicalEncoder{active: make(map[canonicalVisit]struct{})}
	if err := encoder.encode(reflect.ValueOf(value)); err != nil {
		return nil, err
	}
	return append([]byte(nil), encoder.buf.Bytes()...), nil
}

// DigestOf returns the SHA-256 digest of CanonicalBytes(value).
func DigestOf(value any) (Digest, error) {
	data, err := CanonicalBytes(value)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

type canonicalEncoder struct {
	buf    bytes.Buffer
	active map[canonicalVisit]struct{}
}
type canonicalVisit struct {
	typ  reflect.Type
	kind reflect.Kind
	ptr  uintptr
}
