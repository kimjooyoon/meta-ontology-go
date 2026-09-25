package cache

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
)

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
