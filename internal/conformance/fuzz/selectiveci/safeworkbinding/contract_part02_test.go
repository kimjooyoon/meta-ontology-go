package safeworkbinding

import (
	"reflect"
	"testing"
)

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
