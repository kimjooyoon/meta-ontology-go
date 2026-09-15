package cache

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
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
