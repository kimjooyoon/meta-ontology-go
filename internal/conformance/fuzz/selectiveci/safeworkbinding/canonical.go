package safeworkbinding

import "encoding/binary"

type frameTag byte

const (
	frameTagString       frameTag = 0x01
	frameTagStableID     frameTag = 0x02
	frameTagDigest       frameTag = 0x03
	frameTagLegacyWorkID frameTag = 0x04
	frameTagEnum         frameTag = 0x05
	frameTagBool         frameTag = 0x06
	frameTagReasonList   frameTag = 0x07
	frameTagU64          frameTag = 0x08
)

type frameField struct {
	name  string
	tag   frameTag
	value []byte
}

func appendU64BE(out []byte, value uint64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	return append(out, encoded[:]...)
}

func encodeFrame(domain string, fields []frameField) []byte {
	out := appendU64BE(nil, uint64(len(domain)))
	out = append(out, domain...)
	out = appendU64BE(out, uint64(len(fields)))
	for _, field := range fields {
		out = appendU64BE(out, uint64(len(field.name)))
		out = append(out, field.name...)
		out = append(out, byte(field.tag))
		out = appendU64BE(out, uint64(len(field.value)))
		out = append(out, field.value...)
	}
	return out
}

func encodeStringField(name, value string) frameField {
	return frameField{name: name, tag: frameTagString, value: []byte(value)}
}

func encodeStableIDField(name string, value StableID) frameField {
	return frameField{name: name, tag: frameTagStableID, value: []byte(value)}
}

func encodeDigestField(name string, value Digest) frameField {
	return frameField{name: name, tag: frameTagDigest, value: []byte(value)}
}

func encodeLegacyWorkIDField(name string, value LegacyWorkID) frameField {
	return frameField{name: name, tag: frameTagLegacyWorkID, value: []byte(value)}
}

func encodeEnumField(name string, spelling []byte) frameField {
	value := append([]byte(nil), spelling...)
	return frameField{name: name, tag: frameTagEnum, value: value}
}

func encodeBoolField(name string, value bool) frameField {
	encoded := byte(0)
	if value {
		encoded = 1
	}
	return frameField{name: name, tag: frameTagBool, value: []byte{encoded}}
}

func encodeListField(name string, values [][]byte) frameField {
	encoded := appendU64BE(nil, uint64(len(values)))
	for _, value := range values {
		encoded = appendU64BE(encoded, uint64(len(value)))
		encoded = append(encoded, value...)
	}
	return frameField{name: name, tag: frameTagReasonList, value: encoded}
}
