package cache

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (e *canonicalEncoder) encodeBool(value reflect.Value) error {
	e.buf.WriteByte(tagBool)
	e.writeType(value.Type())
	if value.Bool() {
		e.buf.WriteByte(1)
	} else {
		e.buf.WriteByte(0)
	}
	return nil
}

func (e *canonicalEncoder) encodeString(value reflect.Value) error {
	e.buf.WriteByte(tagString)
	e.writeType(value.Type())
	e.writeString(value.String())
	return nil
}

func (e *canonicalEncoder) encodeSigned(value reflect.Value) error {
	e.buf.WriteByte(tagSigned)
	e.writeType(value.Type())
	e.writeString(strconv.FormatInt(value.Int(), 10))
	return nil
}

func (e *canonicalEncoder) encodeUnsigned(value reflect.Value) error {
	e.buf.WriteByte(tagUnsigned)
	e.writeType(value.Type())
	e.writeString(strconv.FormatUint(value.Uint(), 10))
	return nil
}

func (e *canonicalEncoder) encodeFloat(value reflect.Value) error {
	e.buf.WriteByte(tagFloat)
	e.writeType(value.Type())
	if value.Kind() == reflect.Float32 {
		var encoded [4]byte
		binary.BigEndian.PutUint32(encoded[:], math.Float32bits(float32(value.Float())))
		e.buf.Write(encoded[:])
		return nil
	}
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], math.Float64bits(value.Float()))
	e.buf.Write(encoded[:])
	return nil
}

func (e *canonicalEncoder) encodeTime(value reflect.Value) error {
	encoded, err := value.Interface().(time.Time).MarshalText()
	if err != nil {
		return fmt.Errorf("canonical time: %w", err)
	}
	e.buf.WriteByte(tagTime)
	e.writeType(value.Type())
	e.writeBytes(encoded)
	return nil
}

func (e *canonicalEncoder) encodePointer(value reflect.Value) error {
	if value.IsNil() {
		e.buf.WriteByte(tagPointer)
		e.writeType(value.Type())
		e.buf.WriteByte(0)
		return nil
	}
	leave, err := e.enter(value)
	if err != nil {
		return err
	}
	defer leave()
	e.buf.WriteByte(tagPointer)
	e.writeType(value.Type())
	e.buf.WriteByte(1)
	return e.encode(value.Elem())
}

func (e *canonicalEncoder) encodeSlice(value reflect.Value) error {
	if value.IsNil() {
		e.buf.WriteByte(tagSlice)
		e.writeType(value.Type())
		e.writeUvarint(0)
		e.buf.WriteByte(0)
		return nil
	}
	leave, err := e.enter(value)
	if err != nil {
		return err
	}
	defer leave()
	e.buf.WriteByte(tagSlice)
	e.writeType(value.Type())
	e.writeUvarint(uint64(value.Len()))
	e.buf.WriteByte(1)
	for i := 0; i < value.Len(); i++ {
		if err := e.encode(value.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *canonicalEncoder) encodeArray(value reflect.Value) error {
	e.buf.WriteByte(tagArray)
	e.writeType(value.Type())
	e.writeUvarint(uint64(value.Len()))
	for i := 0; i < value.Len(); i++ {
		if err := e.encode(value.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *canonicalEncoder) encodeMap(value reflect.Value) error {
	if value.IsNil() {
		e.writeMapHeader(value, 0, false)
		return nil
	}
	leave, err := e.enter(value)
	if err != nil {
		return err
	}
	defer leave()
	entries := make([]canonicalMapEntry, 0, value.Len())
	iter := value.MapRange()
	for iter.Next() {
		keyBytes, err := CanonicalBytes(iter.Key().Interface())
		if err != nil {
			return fmt.Errorf("canonical map key: %w", err)
		}
		entries = append(entries, canonicalMapEntry{key: keyBytes, value: iter.Value()})
	}
	sort.Slice(entries, func(i, j int) bool {
		return bytes.Compare(entries[i].key, entries[j].key) < 0
	})
	for i := 1; i < len(entries); i++ {
		if bytes.Equal(entries[i-1].key, entries[i].key) {
			return fmt.Errorf("canonical map has duplicate key encoding")
		}
	}
	e.writeMapHeader(value, len(entries), true)
	for _, entry := range entries {
		e.writeBytes(entry.key)
		if err := e.encode(entry.value); err != nil {
			return err
		}
	}
	return nil
}

func (e *canonicalEncoder) writeMapHeader(value reflect.Value, length int, present bool) {
	e.buf.WriteByte(tagMap)
	e.writeType(value.Type())
	e.writeUvarint(uint64(length))
	if present {
		e.buf.WriteByte(1)
	} else {
		e.buf.WriteByte(0)
	}
}

func (e *canonicalEncoder) encodeStruct(value reflect.Value) error {
	fields, err := canonicalFields(value.Type())
	if err != nil {
		return err
	}
	e.buf.WriteByte(tagStruct)
	e.writeType(value.Type())
	e.writeUvarint(uint64(len(fields)))
	for _, field := range fields {
		e.writeString(field.name)
		if err := e.encode(value.Field(field.index)); err != nil {
			return fmt.Errorf("canonical field %s: %w", field.name, err)
		}
	}
	return nil
}

type canonicalMapEntry struct {
	key   []byte
	value reflect.Value
}

type canonicalField struct {
	name  string
	index int
}

func canonicalFields(typ reflect.Type) ([]canonicalField, error) {
	fields := make([]canonicalField, 0, typ.NumField())
	seen := make(map[string]struct{}, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			return nil, fmt.Errorf("unsupported unexported field %s.%s", typ, field.Name)
		}
		name := field.Name
		if tag, ok := field.Tag.Lookup("json"); ok {
			name = strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("duplicate canonical field name %q in %s", name, typ)
		}
		seen[name] = struct{}{}
		fields = append(fields, canonicalField{name: name, index: i})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].name < fields[j].name })
	return fields, nil
}

func typeDescriptor(typ reflect.Type) string {
	if typ.PkgPath() != "" && typ.Name() != "" {
		return typ.PkgPath() + "." + typ.Name()
	}
	return typ.String()
}
