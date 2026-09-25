package cache

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

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
