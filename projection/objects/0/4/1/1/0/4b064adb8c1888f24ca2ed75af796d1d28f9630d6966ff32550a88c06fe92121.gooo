package safeworkbinding

import (
	"reflect"
	"testing"
)

func TestA11Declarations(t *testing.T) {
	check(t, SafeWorkBindingSchemaV1 == "gooo/safe-work-binding/v1", "schema")
	check(t, reflect.TypeFor[LegacyWorkID]().Kind() == reflect.String, "legacy type")
	check(t, reflect.TypeFor[Digest]().Kind() == reflect.String, "digest type")
	check(t, reflect.TypeFor[StableID]().Kind() == reflect.String, "stable ID type")
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
	checkFields(t, reflect.TypeFor[SafeWorkBinding](), []fieldSpec{
		{"Schema", reflect.TypeFor[string](), "schema"},
		{"TaskID", reflect.TypeFor[StableID](), "task_id"},
		{"PathID", reflect.TypeFor[StableID](), "path_id"},
		{"ObligationID", reflect.TypeFor[StableID](), "obligation_id"},
		{"SourceSnapshotDigest", reflect.TypeFor[Digest](), "source_snapshot_digest"},
		{"SemanticSnapshotDigest", reflect.TypeFor[Digest](), "semantic_snapshot_digest"},
		{"PolicyDigest", reflect.TypeFor[Digest](), "policy_digest"},
		{"RegistryDigest", reflect.TypeFor[Digest](), "registry_digest"},
		{"ToolchainOptionsDigest", reflect.TypeFor[Digest](), "toolchain_options_digest"},
		{"BindingDigest", reflect.TypeFor[Digest](), "binding_digest"},
	})
	checkFields(t, reflect.TypeFor[ParseResult](), []fieldSpec{
		{"Decision", reflect.TypeFor[Decision](), ""},
		{"Reason", reflect.TypeFor[Reason](), ""},
		{"Faults", reflect.TypeFor[[]Reason](), ""},
		{"FullSuiteRequired", reflect.TypeOf(false), ""},
		{"ExecutionAuthorized", reflect.TypeOf(false), ""},
		{"EnforcementEffect", reflect.TypeFor[EnforcementEffect](), ""},
		{"ResultDigest", reflect.TypeFor[Digest](), ""},
		{"ReplayDigest", reflect.TypeFor[Digest](), ""},
	})
}
