package cache

import (
	"reflect"
)

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
