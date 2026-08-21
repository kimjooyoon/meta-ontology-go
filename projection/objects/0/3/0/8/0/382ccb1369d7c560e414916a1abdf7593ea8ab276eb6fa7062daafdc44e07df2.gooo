package eligibilityregistry

import (
	"encoding/binary"
)

type frameField struct {
	name  string
	tag   byte
	value []byte
}

const (
	frameTagString     byte = 0x01
	frameTagStableID   byte = 0x02
	frameTagDigest     byte = 0x03
	frameTagEnum       byte = 0x05
	frameTagBool       byte = 0x06
	frameTagReasonList byte = 0x07
	frameTagU64        byte = 0x08
	frameTagRecordList byte = 0x09
)

func encodeFrame(domain string, fields []frameField) []byte {
	frame := appendU64BE(nil, uint64(len(domain)))
	frame = append(frame, domain...)
	frame = appendU64BE(frame, uint64(len(fields)))
	for _, field := range fields {
		encoded := encodeField(field)
		if encoded == nil {
			return nil
		}
		frame = append(frame, encoded...)
	}
	return frame
}
func encodeField(field frameField) []byte {
	if field.value == nil {
		return nil
	}
	switch field.tag {
	case frameTagString, frameTagStableID, frameTagDigest, frameTagEnum:
	case frameTagBool, frameTagReasonList, frameTagU64, frameTagRecordList:
	default:
		return nil
	}
	encoded := appendU64BE(nil, uint64(len(field.name)))
	encoded = append(encoded, field.name...)
	encoded = append(encoded, field.tag)
	encoded = appendU64BE(encoded, uint64(len(field.value)))
	return append(encoded, field.value...)
}
func appendU64BE(dst []byte, value uint64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	return append(dst, encoded[:]...)
}
func recordListValue(records [][]byte) []byte {
	value := appendU64BE(nil, uint64(len(records)))
	for _, record := range records {
		if record == nil {
			return nil
		}
		value = appendU64BE(value, uint64(len(record)))
		value = append(value, record...)
	}
	return value
}
