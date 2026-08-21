package safeworkbinding

import (
	"encoding/hex"
	"testing"
)

func TestEnumSpelling_NoEffect(t *testing.T) {
	expected := []struct {
		value    EnforcementEffect
		spelling string
	}{
		{EnforcementEffectNoEffect, "NO_EFFECT"},
	}
	for _, vector := range expected {
		field := encodeEnumField("effect", []byte(vector.spelling))
		check(t, field.tag == frameTagEnum, "effect tag")
		check(t, uint8(vector.value) == 0, "effect value")
		check(t, len(field.value) == 9, "effect length")
		check(t, hex.EncodeToString(field.value) == "4e4f5f454646454354", "effect spelling")
	}
}
func TestA11PrimitiveFrames(t *testing.T) {
	for _, vector := range []struct {
		value uint64
		want  string
	}{
		{0, "0000000000000000"},
		{1, "0000000000000001"},
		{^uint64(0), "ffffffffffffffff"},
	} {
		check(t, hex.EncodeToString(appendU64BE(nil, vector.value)) == vector.want, "u64")
	}
	for i, tag := range []frameTag{
		frameTagString,
		frameTagStableID,
		frameTagDigest,
		frameTagLegacyWorkID,
		frameTagEnum,
		frameTagBool,
		frameTagReasonList,
		frameTagU64,
	} {
		check(t, byte(tag) == byte(i+1), "tag")
	}
	checkField(t, encodeStringField("x", "abc"), frameTagString, []byte("abc"))
	checkField(t, encodeStableIDField("x", StableID("abc")), frameTagStableID, []byte("abc"))
	checkField(t, encodeDigestField("x", Digest("abc")), frameTagDigest, []byte("abc"))
	checkField(t, encodeLegacyWorkIDField("x", LegacyWorkID("abc")), frameTagLegacyWorkID, []byte("abc"))
	checkField(t, encodeEnumField("x", []byte("PASS")), frameTagEnum, []byte("PASS"))
	checkField(t, encodeBoolField("x", false), frameTagBool, []byte{0})
	checkField(t, encodeBoolField("x", true), frameTagBool, []byte{1})
	checkField(t, encodeListField("x", [][]byte{{1}, {2, 3}}), frameTagReasonList, []byte{
		0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 2, 2, 3,
	})
	nilList := hex.EncodeToString(encodeListField("x", nil).value)
	emptyList := hex.EncodeToString(encodeListField("x", [][]byte{}).value)
	check(t, nilList == emptyList, "nil list")
	frame := encodeFrame("d\x00", []frameField{{name: "x", tag: frameTagU64, value: []byte{1, 2}}})
	frameHex := hex.EncodeToString(frame)
	check(t, frameHex == "0000000000000002640000000000000000010000000000000001780800000000000000020102",
		"frame layout")
	legacy := encodeFrame("p", []frameField{encodeLegacyWorkIDField("legacy_work_id", LegacyWorkID("abc"))})
	legacyHex := hex.EncodeToString(legacy)
	legacyWant := "0000000000000001700000000000000001000000000000000e6c65676163795f776f726b5f6964" +
		"040000000000000003616263"
	check(t, legacyHex == legacyWant,
		"legacy frame")
}
