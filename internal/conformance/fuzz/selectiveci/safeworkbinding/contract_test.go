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
	}{
		{DecisionPass, "PASS"},
		{DecisionUnknown, "UNKNOWN"},
		{DecisionFailClosed, "FAIL_CLOSED"},
	}
	for i, expected := range decisions {
		check(t, uint8(expected.value) == uint8(i), "decision value")
		check(t, string(encodeEnumField("decision", []byte(expected.spelling)).value) == expected.spelling,
			"decision spelling")
	}
	reasons := []struct {
		value    Reason
		spelling string
	}{
		{ReasonNone, "NONE"},
		{ReasonRequiredInputMissing, "REQUIRED_INPUT_MISSING"},
		{ReasonInvalidUTF8, "INVALID_UTF8"},
		{ReasonBOMForbidden, "BOM_FORBIDDEN"},
		{ReasonInvalidJSON, "INVALID_JSON"},
		{ReasonTrailingValue, "TRAILING_VALUE"},
		{ReasonDuplicateKey, "DUPLICATE_KEY"},
		{ReasonUnknownField, "UNKNOWN_FIELD"},
		{ReasonNullValue, "NULL_VALUE"},
		{ReasonEmptyValue, "EMPTY_VALUE"},
		{ReasonInvalidSchema, "INVALID_SCHEMA"},
		{ReasonInvalidStableID, "INVALID_STABLE_ID"},
		{ReasonInvalidDigest, "INVALID_DIGEST"},
		{ReasonBindingDigestMismatch, "BINDING_DIGEST_MISMATCH"},
	}
	for i, expected := range reasons {
		check(t, uint8(expected.value) == uint8(i), "reason value")
		check(t, string(encodeEnumField("reason", []byte(expected.spelling)).value) == expected.spelling,
			"reason spelling")
	}
	check(t, uint8(EnforcementEffectNoEffect) == 0, "effect value")
	checkFields(t, reflect.TypeOf(SafeWorkBinding{}), []fieldSpec{
		{"Schema", reflect.TypeOf(""), "schema"},
		{"TaskID", reflect.TypeOf(StableID("")), "task_id"},
		{"PathID", reflect.TypeOf(StableID("")), "path_id"},
		{"ObligationID", reflect.TypeOf(StableID("")), "obligation_id"},
		{"SourceSnapshotDigest", reflect.TypeOf(Digest("")), "source_snapshot_digest"},
		{"SemanticSnapshotDigest", reflect.TypeOf(Digest("")), "semantic_snapshot_digest"},
		{"PolicyDigest", reflect.TypeOf(Digest("")), "policy_digest"},
		{"RegistryDigest", reflect.TypeOf(Digest("")), "registry_digest"},
		{"ToolchainOptionsDigest", reflect.TypeOf(Digest("")), "toolchain_options_digest"},
		{"BindingDigest", reflect.TypeOf(Digest("")), "binding_digest"},
	})
	checkFields(t, reflect.TypeOf(ParseResult{}), []fieldSpec{
		{"Decision", reflect.TypeOf(Decision(0)), ""},
		{"Reason", reflect.TypeOf(Reason(0)), ""},
		{"Faults", reflect.TypeOf([]Reason(nil)), ""},
		{"FullSuiteRequired", reflect.TypeOf(false), ""},
		{"ExecutionAuthorized", reflect.TypeOf(false), ""},
		{"EnforcementEffect", reflect.TypeOf(EnforcementEffect(0)), ""},
		{"ResultDigest", reflect.TypeOf(Digest("")), ""},
		{"ReplayDigest", reflect.TypeOf(Digest("")), ""},
	})
}

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

const (
	digestBase          = "sha256:dc6dbe157ede5924b61676bfdcd4151cd6f73a51b7eefda674cca3d6d169a5cb"
	digestPass          = "sha256:b076bd08e4b82b4b2aeb78f3f8a3e12931ae8f8a01f734cb2a97bca15753a342"
	digestReplay        = "sha256:4564f79ab93fadeb3221f837e38dfdce43e1c653a89086283aaceb09d923080d"
	digestPath          = "sha256:c262c10f1652d0d3fd5fa605166d565a44261b834ffcffe876f2bb785f4bcd51"
	digestRegistry      = "sha256:23477b01e3d477d6ba5702adca8ad77a2cd643748b29e20ac4a7fda0b020673e"
	digestMissing       = "sha256:3b4256d3ab6d60db4caa093cc8306819c9a6fb9fb987927fab8bff841a0690c8"
	digestDuplicate     = "sha256:bac3126c6e354dfb38abf69508605e50226f186d25b7381f8eb3215f33e7fda0"
	digestExpectedLabel = "sha256:1ada140e4d914ab1ab15570deb11e423a7e454b81c2bbf84512454b642ecfa02"
	digestMismatch      = "sha256:c61d5823d85939c5f7645d6f2fe2049fe724bfec3fc12c011e9532ccc41c6b10"
	digestBOMMixed      = "sha256:49d26986dde23834ba79d1064aa577860dfe47a79bff2ab698d385660db1ac8c"
)

func repeatedDigest(digit byte) Digest {
	value := make([]byte, 64)
	for i := range value {
		value[i] = digit
	}
	return Digest("sha256:" + string(value))
}

func baseBindingForDigest() SafeWorkBinding {
	return SafeWorkBinding{
		Schema:                 SafeWorkBindingSchemaV1,
		TaskID:                 "billing://task/pay",
		PathID:                 "billing://path/pay",
		ObligationID:           "billing://obligation/pay",
		SourceSnapshotDigest:   repeatedDigest('1'),
		SemanticSnapshotDigest: repeatedDigest('2'),
		PolicyDigest:           repeatedDigest('3'),
		RegistryDigest:         repeatedDigest('4'),
		ToolchainOptionsDigest: repeatedDigest('5'),
	}
}

func passResultForDigest() ParseResult {
	return ParseResult{Decision: DecisionPass, Reason: ReasonNone, Faults: []Reason{},
		EnforcementEffect: EnforcementEffectNoEffect}
}

func fixtureResult(decision Decision, reason Reason, fullSuite bool) ParseResult {
	return ParseResult{Decision: decision, Reason: reason, Faults: []Reason{reason},
		FullSuiteRequired: fullSuite, EnforcementEffect: EnforcementEffectNoEffect}
}

func checkResultDigest(t *testing.T, result ParseResult, want Digest) {
	t.Helper()
	frame, frameOK := resultFrame(result)
	check(t, frameOK, "result frame")
	check(t, len(frame) > 0, "result frame bytes")
	digest, digestOK := resultDigest(result)
	check(t, digestOK, "result digest")
	check(t, digest == want, "result vector")
}

func TestBindingDigest_Base(t *testing.T) {
	binding := baseBindingForDigest()
	check(t, len(bindingFrame(binding)) == 772, "binding frame length")
	check(t, bindingDigest(binding) == digestBase, "binding digest")
}

func TestResultDigest_PASS(t *testing.T) {
	result := passResultForDigest()
	frame, ok := resultFrame(result)
	check(t, ok, "pass result frame")
	check(t, len(frame) == 255, "pass result frame length")
	digest, ok := resultDigest(result)
	check(t, ok, "pass result digest")
	check(t, digest == digestPass, "pass result vector")
}

func TestReplayDigest_PASS(t *testing.T) {
	binding := bindingDigest(baseBindingForDigest())
	result, ok := resultDigest(passResultForDigest())
	check(t, ok, "pass replay input")
	frame := replayFrame(binding, result)
	check(t, len(frame) == 252, "replay frame length")
	check(t, replayDigest(binding, result) == digestReplay, "replay vector")
	check(t, replayDigest(binding, result) == replayDigest(binding, result), "replay repeat")
}

func TestBindingDigest_PathIDMutation(t *testing.T) {
	binding := baseBindingForDigest()
	binding.PathID = "billing://path/pay-v2"
	check(t, bindingDigest(binding) == digestPath, "path mutation vector")
}

func TestBindingDigest_RegistryMutation(t *testing.T) {
	binding := baseBindingForDigest()
	binding.RegistryDigest = repeatedDigest('9')
	check(t, bindingDigest(binding) == digestRegistry, "registry mutation vector")
}

func TestCanonicalResult_MissingBindingDigestFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionUnknown, ReasonRequiredInputMissing, true), digestMissing)
}

func TestCanonicalResult_DuplicateKeyFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonDuplicateKey, false), digestDuplicate)
}

func TestCanonicalResult_ExpectedLabelFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonUnknownField, false), digestExpectedLabel)
}

func TestCanonicalResult_BindingDigestMismatchFixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonBindingDigestMismatch, false), digestMismatch)
}

func TestCanonicalResult_BOMForbiddenMixedInvalidUTF8Fixture(t *testing.T) {
	checkResultDigest(t, fixtureResult(DecisionFailClosed, ReasonBOMForbidden, false), digestBOMMixed)
}

func TestCanonicalResult_SelfExclusion(t *testing.T) {
	binding := baseBindingForDigest()
	wantBinding := bindingDigest(binding)
	binding.BindingDigest = "sha256:mutated"
	check(t, bindingDigest(binding) == wantBinding, "binding self exclusion")
	result := passResultForDigest()
	wantResult, ok := resultDigest(result)
	check(t, ok, "result self input")
	result.ResultDigest = "sha256:mutated"
	gotResult, ok := resultDigest(result)
	check(t, ok, "result self digest")
	check(t, gotResult == wantResult, "result self exclusion")
}

func TestCanonicalResult_InvalidEnumValues(t *testing.T) {
	_, ok := decisionField("decision", Decision(255))
	check(t, !ok, "invalid decision")
	_, ok = reasonField("reason", Reason(255))
	check(t, !ok, "invalid reason")
	_, ok = enforcementEffectField("effect", EnforcementEffect(255))
	check(t, !ok, "invalid effect")
	_, ok = reasonListField("faults", []Reason{Reason(255)})
	check(t, !ok, "invalid reason list")
	_, ok = resultFrame(ParseResult{Decision: Decision(255)})
	check(t, !ok, "invalid result")
}

func TestBindingDigest_ConstructionOrderAndListInvariance(t *testing.T) {
	binding := baseBindingForDigest()
	permuted := SafeWorkBinding{
		ToolchainOptionsDigest: binding.ToolchainOptionsDigest,
		RegistryDigest:         binding.RegistryDigest,
		PolicyDigest:           binding.PolicyDigest,
		SemanticSnapshotDigest: binding.SemanticSnapshotDigest,
		SourceSnapshotDigest:   binding.SourceSnapshotDigest,
		ObligationID:           binding.ObligationID,
		PathID:                 binding.PathID,
		TaskID:                 binding.TaskID,
		Schema:                 binding.Schema,
	}
	check(t, bindingDigest(binding) == bindingDigest(permuted), "binding construction order")
	nilResult := passResultForDigest()
	nilResult.Faults = nil
	emptyResult := passResultForDigest()
	nilDigest, nilOK := resultDigest(nilResult)
	emptyDigest, emptyOK := resultDigest(emptyResult)
	check(t, nilOK, "nil result list")
	check(t, emptyOK, "empty result list")
	check(t, nilDigest == emptyDigest, "result list invariance")
}

func TestBindingDigest_GovernedFieldMutations(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*SafeWorkBinding)
	}{
		{
			"schema",
			func(binding *SafeWorkBinding) { binding.Schema += "-v2" },
		},
		{
			"task",
			func(binding *SafeWorkBinding) { binding.TaskID += "-v2" },
		},
		{
			"path",
			func(binding *SafeWorkBinding) { binding.PathID += "-v2" },
		},
		{
			"obligation",
			func(binding *SafeWorkBinding) { binding.ObligationID += "-v2" },
		},
		{
			"source",
			func(binding *SafeWorkBinding) { binding.SourceSnapshotDigest += "-v2" },
		},
		{
			"semantic",
			func(binding *SafeWorkBinding) { binding.SemanticSnapshotDigest += "-v2" },
		},
		{
			"policy",
			func(binding *SafeWorkBinding) { binding.PolicyDigest += "-v2" },
		},
		{
			"registry",
			func(binding *SafeWorkBinding) { binding.RegistryDigest += "-v2" },
		},
		{
			"toolchain",
			func(binding *SafeWorkBinding) { binding.ToolchainOptionsDigest += "-v2" },
		},
	}
	base := bindingDigest(baseBindingForDigest())
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			binding := baseBindingForDigest()
			mutation.mutate(&binding)
			check(t, bindingDigest(binding) != base, "governed mutation")
		})
	}
}
