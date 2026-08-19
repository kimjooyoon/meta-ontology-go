package eligibilityregistry

import (
	"testing"
)

func TestU64AndListVectors(t *testing.T) {
	requireFrameBytes(t, "u64 zero", appendU64BE(nil, 0), []byte{0, 0, 0, 0, 0, 0, 0, 0})
	requireFrameBytes(t, "u64 one", appendU64BE(nil, 1), []byte{0, 0, 0, 0, 0, 0, 0, 1})
	requireFrameBytes(t, "u64 max", appendU64BE(nil, ^uint64(0)), []byte{255, 255, 255, 255, 255, 255, 255, 255})
	boolFalse := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x06, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	requireFrameBytes(t, "bool false", encodeField(frameField{tag: frameTagBool, value: []byte{0}}), boolFalse)
	boolTrue := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x06, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	requireFrameBytes(t, "bool true", encodeField(frameField{tag: frameTagBool, value: []byte{1}}), boolTrue)
	reasonList := frameField{tag: frameTagReasonList, value: []byte{0, 0, 0, 0, 0, 0, 0, 0}}
	reasonListFrame := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x07, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0}
	requireFrameBytes(t, "reason list", encodeField(reasonList), reasonListFrame)
	u64Field := frameField{tag: frameTagU64, value: []byte{0, 0, 0, 0, 0, 0, 0, 1}}
	u64Frame := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x08, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 1}
	requireFrameBytes(t, "u64 field", encodeField(u64Field), u64Frame)
	recordField := frameField{tag: frameTagRecordList, value: []byte{0, 0, 0, 0, 0, 0, 0, 0}}
	recordFrame := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x09, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0}
	requireFrameBytes(t, "record-list field", encodeField(recordField), recordFrame)
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
