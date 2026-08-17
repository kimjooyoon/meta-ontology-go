package safeworkbinding

import (
	"encoding/hex"
	"reflect"
	"testing"
)

type fieldSpec struct {
	name string
	typ  reflect.Type
	tag  string
}

func checkFields(t *testing.T, typ reflect.Type, want []fieldSpec) {
	t.Helper()
	if typ.NumField() != len(want) {
		t.Fatalf("field count: got %d want %d", typ.NumField(), len(want))
	}
	for i, expected := range want {
		field := typ.Field(i)
		if field.Name != expected.name || field.Type != expected.typ {
			t.Fatalf("field %d: got %s %s", i, field.Name, field.Type)
		}
		if field.Tag.Get("json") != expected.tag {
			t.Fatalf("field %s: json tag %q", field.Name, field.Tag.Get("json"))
		}
	}
}

func check(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatal(message)
	}
}

func checkField(t *testing.T, got frameField, tag frameTag, value []byte) {
	t.Helper()
	check(t, got.tag == tag, "field tag")
	check(t, hex.EncodeToString(got.value) == hex.EncodeToString(value), "field value")
}

func TestA11Declarations(t *testing.T) {
	check(t, SafeWorkBindingSchemaV1 == "gooo/safe-work-binding/v1", "schema")
	check(t, reflect.TypeOf(LegacyWorkID("")).Kind() == reflect.String, "legacy type")
	check(t, reflect.TypeOf(Digest("")).Kind() == reflect.String, "digest type")
	check(t, reflect.TypeOf(StableID("")).Kind() == reflect.String, "stable ID type")
	decisions := []struct {
		value    Decision
		spelling string
	}{{DecisionPass, "PASS"}, {DecisionUnknown, "UNKNOWN"}, {DecisionFailClosed, "FAIL_CLOSED"}}
	for i, expected := range decisions {
		check(t, uint8(expected.value) == uint8(i), "decision value")
		check(t, string(encodeEnumField("decision", []byte(expected.spelling)).value) == expected.spelling,
			"decision spelling")
	}
	reasons := []struct {
		value    Reason
		spelling string
	}{
		{ReasonNone, "NONE"}, {ReasonRequiredInputMissing, "REQUIRED_INPUT_MISSING"},
		{ReasonInvalidUTF8, "INVALID_UTF8"}, {ReasonBOMForbidden, "BOM_FORBIDDEN"},
		{ReasonInvalidJSON, "INVALID_JSON"}, {ReasonTrailingValue, "TRAILING_VALUE"},
		{ReasonDuplicateKey, "DUPLICATE_KEY"}, {ReasonUnknownField, "UNKNOWN_FIELD"},
		{ReasonNullValue, "NULL_VALUE"}, {ReasonEmptyValue, "EMPTY_VALUE"},
		{ReasonInvalidSchema, "INVALID_SCHEMA"}, {ReasonInvalidStableID, "INVALID_STABLE_ID"},
		{ReasonInvalidDigest, "INVALID_DIGEST"}, {ReasonBindingDigestMismatch, "BINDING_DIGEST_MISMATCH"},
	}
	for i, expected := range reasons {
		check(t, uint8(expected.value) == uint8(i), "reason value")
		check(t, string(encodeEnumField("reason", []byte(expected.spelling)).value) == expected.spelling,
			"reason spelling")
	}
	check(t, uint8(EnforcementEffectNoEffect) == 0, "effect value")
	checkFields(t, reflect.TypeOf(SafeWorkBinding{}), []fieldSpec{
		{"Schema", reflect.TypeOf(""), "schema"}, {"TaskID", reflect.TypeOf(StableID("")), "task_id"},
		{"PathID", reflect.TypeOf(StableID("")), "path_id"}, {"ObligationID", reflect.TypeOf(StableID("")), "obligation_id"},
		{"SourceSnapshotDigest", reflect.TypeOf(Digest("")), "source_snapshot_digest"},
		{"SemanticSnapshotDigest", reflect.TypeOf(Digest("")), "semantic_snapshot_digest"},
		{"PolicyDigest", reflect.TypeOf(Digest("")), "policy_digest"},
		{"RegistryDigest", reflect.TypeOf(Digest("")), "registry_digest"},
		{"ToolchainOptionsDigest", reflect.TypeOf(Digest("")), "toolchain_options_digest"},
		{"BindingDigest", reflect.TypeOf(Digest("")), "binding_digest"},
	})
	checkFields(t, reflect.TypeOf(ParseResult{}), []fieldSpec{
		{"Decision", reflect.TypeOf(Decision(0)), ""}, {"Reason", reflect.TypeOf(Reason(0)), ""},
		{"Faults", reflect.TypeOf([]Reason(nil)), ""}, {"FullSuiteRequired", reflect.TypeOf(false), ""},
		{"ExecutionAuthorized", reflect.TypeOf(false), ""}, {"EnforcementEffect", reflect.TypeOf(EnforcementEffect(0)), ""},
		{"ResultDigest", reflect.TypeOf(Digest("")), ""}, {"ReplayDigest", reflect.TypeOf(Digest("")), ""},
	})
}

func TestA11PrimitiveFrames(t *testing.T) {
	for _, vector := range []struct {
		value uint64
		want  string
	}{{0, "0000000000000000"}, {1, "0000000000000001"}, {^uint64(0), "ffffffffffffffff"}} {
		check(t, hex.EncodeToString(appendU64BE(nil, vector.value)) == vector.want, "u64")
	}
	for i, tag := range []frameTag{
		frameTagString, frameTagStableID, frameTagDigest, frameTagLegacyWorkID,
		frameTagEnum, frameTagBool, frameTagReasonList, frameTagU64,
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
