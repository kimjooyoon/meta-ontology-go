package eligibilityregistry

import (
	"bytes"
	"testing"
)

func requireFrameBytes(t *testing.T, name string, got, want []byte) {
	t.Helper()
	requireFrame(t, name, bytes.Equal(got, want))
}

func requireFrame(t *testing.T, name string, condition bool) {
	t.Helper()
	if !condition {
		t.Fatal(name)
	}
}

func TestFrameFieldNilAndEmpty(t *testing.T) {
	requireFrame(t, "nil value", encodeField(frameField{name: "x", tag: frameTagString}) == nil)
	empty := encodeField(frameField{name: "x", tag: frameTagString, value: []byte{}})
	want := []byte{0, 0, 0, 0, 0, 0, 0, 1, 'x', 0x01, 0, 0, 0, 0, 0, 0, 0, 0}
	requireFrameBytes(t, "empty value", empty, want)
	requireFrame(t, "invalid tag", encodeField(frameField{tag: 0x04, value: []byte{}}) == nil)
	nilFields, emptyFields := encodeFrame("d", nil), encodeFrame("d", []frameField{})
	requireFrame(t, "nil fields", nilFields != nil && len(nilFields) == 17)
	requireFrameBytes(t, "empty fields", emptyFields, nilFields)
	requireFrame(t, "nil frame field", encodeFrame("d", []frameField{{tag: frameTagString}}) == nil)
}

func TestEncodeFieldAndFrameVectors(t *testing.T) {
	field := encodeField(frameField{name: "id", tag: frameTagStableID, value: []byte("x")})
	wantField := []byte{0, 0, 0, 0, 0, 0, 0, 2, 'i', 'd', 0x02, 0, 0, 0, 0, 0, 0, 0, 1, 'x'}
	requireFrameBytes(t, "field", field, wantField)
	emptyName := encodeField(frameField{tag: frameTagEnum, value: []byte("PASS")})
	wantEmptyName := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x05, 0, 0, 0, 0, 0, 0, 0, 4, 'P', 'A', 'S', 'S'}
	requireFrameBytes(t, "empty name", emptyName, wantEmptyName)
	frame := encodeFrame("d", []frameField{
		{name: "a", tag: frameTagString, value: []byte("1")},
		{name: "b", tag: frameTagDigest, value: []byte("2")},
	})
	wantFrame := []byte{
		0, 0, 0, 0, 0, 0, 0, 1, 'd', 0, 0, 0, 0, 0, 0, 0, 2,
		0, 0, 0, 0, 0, 0, 0, 1, 'a', 0x01, 0, 0, 0, 0, 0, 0, 0, 1, '1',
		0, 0, 0, 0, 0, 0, 0, 1, 'b', 0x03, 0, 0, 0, 0, 0, 0, 0, 1, '2',
	}
	requireFrameBytes(t, "ordered frame", frame, wantFrame)
}

func TestU64AndListVectors(t *testing.T) {
	requireFrameBytes(t, "u64 zero", appendU64BE(nil, 0), []byte{0, 0, 0, 0, 0, 0, 0, 0})
	requireFrameBytes(t, "u64 one", appendU64BE(nil, 1), []byte{0, 0, 0, 0, 0, 0, 0, 1})
	requireFrameBytes(t, "u64 max", appendU64BE(nil, ^uint64(0)), []byte{255, 255, 255, 255, 255, 255, 255, 255})
	boolFalse := encodeField(frameField{tag: frameTagBool, value: []byte{0}})
	requireFrame(t, "bool false", bytes.HasSuffix(boolFalse, []byte{0}))
	requireFrame(t, "bool true", bytes.HasSuffix(encodeField(frameField{tag: frameTagBool, value: []byte{1}}), []byte{1}))
	reasonList := frameField{tag: frameTagReasonList, value: []byte{0, 0, 0, 0, 0, 0, 0, 0}}
	requireFrame(t, "reason list", encodeField(reasonList) != nil)
	u64Field := frameField{tag: frameTagU64, value: []byte{0, 0, 0, 0, 0, 0, 0, 1}}
	requireFrame(t, "u64 field", encodeField(u64Field) != nil)
	recordField := frameField{tag: frameTagRecordList, value: []byte{0, 0, 0, 0, 0, 0, 0, 0}}
	requireFrame(t, "record-list field", encodeField(recordField) != nil)
	nilList, emptyList := recordListValue(nil), recordListValue([][]byte{})
	requireFrame(t, "non-nil lists", nilList != nil && emptyList != nil)
	requireFrameBytes(t, "nil list", nilList, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	requireFrameBytes(t, "empty list", emptyList, []byte{0, 0, 0, 0, 0, 0, 0, 0})
	nilList[0] = 1
	requireFrame(t, "fresh lists", emptyList[0] == 0)
	requireFrame(t, "nil record", recordListValue([][]byte{nil}) == nil)
	emptyRecord := []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	requireFrameBytes(t, "empty record", recordListValue([][]byte{{}}), emptyRecord)
	record := []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0xAA}
	requireFrameBytes(t, "record", recordListValue([][]byte{{0xAA}}), record)
}

func TestTypedSpellingAndInvalidEnums(t *testing.T) {
	requireFrameBytes(t, "semantic", itemKindSpelling(ItemSemantic), []byte("SEMANTIC"))
	requireFrameBytes(t, "structural", itemKindSpelling(ItemStructural), []byte("STRUCTURAL"))
	requireFrameBytes(t, "business", authorityKindSpelling(AuthorityBusinessDSL), []byte("BUSINESS_DSL"))
	requireFrameBytes(t, "authority semantic", authorityKindSpelling(AuthoritySemanticIR), []byte("SEMANTIC_IR"))
	requireFrameBytes(t, "projection semantic", projectionKindSpelling(ProjectionSemanticIR), []byte("SEMANTIC_IR"))
	requireFrameBytes(t, "generated", projectionKindSpelling(ProjectionGeneratedGo), []byte("GENERATED_GO"))
	requireFrame(t, "invalid item", itemKindSpelling(255) == nil)
	requireFrame(t, "invalid authority", authorityKindSpelling(255) == nil)
	requireFrame(t, "invalid projection", projectionKindSpelling(255) == nil)
	requireFrame(t, "string tag", frameTagString == 0x01)
	requireFrame(t, "stable ID tag", frameTagStableID == 0x02)
	requireFrame(t, "digest tag", frameTagDigest == 0x03)
	requireFrame(t, "enum tag", frameTagEnum == 0x05)
	requireFrame(t, "bool tag", frameTagBool == 0x06)
	requireFrame(t, "reason-list tag", frameTagReasonList == 0x07)
	requireFrame(t, "u64 tag", frameTagU64 == 0x08)
	requireFrame(t, "record-list tag", frameTagRecordList == 0x09)
}

func TestDigestBytesVector(t *testing.T) {
	want := Digest("sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	requireFrame(t, "digest", digestBytes([]byte("abc")) == want)
}
