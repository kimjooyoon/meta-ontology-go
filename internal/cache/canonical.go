package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"
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

const (
	tagNil byte = iota
	tagBool
	tagString
	tagSigned
	tagUnsigned
	tagFloat
	tagTime
	tagPointer
	tagSlice
	tagArray
	tagMap
	tagStruct
)

func (e *canonicalEncoder) encode(value reflect.Value) error {
	if !value.IsValid() {
		e.buf.WriteByte(tagNil)
		return nil
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			e.buf.WriteByte(tagNil)
			return nil
		}
		return e.encode(value.Elem())
	}
	if value.Type() == reflect.TypeOf(time.Time{}) {
		return e.encodeTime(value)
	}
	switch value.Kind() {
	case reflect.Bool:
		return e.encodeBool(value)
	case reflect.String:
		return e.encodeString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.encodeSigned(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return e.encodeUnsigned(value)
	case reflect.Float32, reflect.Float64:
		return e.encodeFloat(value)
	case reflect.Pointer:
		return e.encodePointer(value)
	case reflect.Slice:
		return e.encodeSlice(value)
	case reflect.Array:
		return e.encodeArray(value)
	case reflect.Map:
		return e.encodeMap(value)
	case reflect.Struct:
		return e.encodeStruct(value)
	default:
		return fmt.Errorf("unsupported canonical value of type %s", value.Type())
	}
}

func (e *canonicalEncoder) enter(value reflect.Value) (func(), error) {
	ptr := uintptr(value.UnsafePointer())
	if ptr == 0 {
		return func() {}, nil
	}
	visit := canonicalVisit{typ: value.Type(), kind: value.Kind(), ptr: ptr}
	if _, exists := e.active[visit]; exists {
		return nil, fmt.Errorf("cyclic value of type %s cannot be canonicalized", value.Type())
	}
	e.active[visit] = struct{}{}
	return func() { delete(e.active, visit) }, nil
}

func (e *canonicalEncoder) writeType(typ reflect.Type) {
	e.writeString(typeDescriptor(typ))
}

func (e *canonicalEncoder) writeString(value string) {
	e.writeBytes([]byte(value))
}

func (e *canonicalEncoder) writeBytes(value []byte) {
	e.writeUvarint(uint64(len(value)))
	e.buf.Write(value)
}

func (e *canonicalEncoder) writeUvarint(value uint64) {
	var encoded [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(encoded[:], value)
	e.buf.Write(encoded[:n])
}
