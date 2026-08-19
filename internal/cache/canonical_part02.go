package cache

import (
	"fmt"
	"reflect"
	"time"
)

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
