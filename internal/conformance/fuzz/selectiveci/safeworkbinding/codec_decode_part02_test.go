package safeworkbinding

import (
	"strings"
	"testing"
)

func requireDecodePass(t *testing.T, input []byte, want SafeWorkBinding) ParseResult {
	t.Helper()
	binding, result := DecodeJSON(input)
	check(t, binding == want, "pass binding")
	check(t, result.Decision == DecisionPass, "pass decision")
	check(t, result.Reason == ReasonNone, "pass reason")
	check(t, result.Faults != nil && len(result.Faults) == 0, "pass faults")
	check(t, !result.FullSuiteRequired, "pass suite")
	check(t, result.ResultDigest == digestPass, "pass result digest")
	check(t, result.ReplayDigest == digestReplay, "pass replay")
	check(t, !result.ExecutionAuthorized, "pass authorization")
	check(t, result.EnforcementEffect == EnforcementEffectNoEffect, "pass effect")
	return result
}
func checkGovernedDecode(t *testing.T, field, value string, digest Digest) {
	t.Helper()
	binding := decodeBaseBinding()
	mutateEnvelopeBinding(&binding, field, value)
	overrides := map[string]string{field: `"` + value + `"`}
	requireDecodeFault(t, decodeDocument(nil, overrides), ReasonBindingDigestMismatch)
	binding.BindingDigest = digest
	overrides["binding_digest"] = `"` + string(digest) + `"`
	decoded, result := DecodeJSON(decodeDocument(nil, overrides))
	check(t, decoded == binding, "governed binding")
	check(t, result.Decision == DecisionPass && result.Reason == ReasonNone, "governed pass")
}
func TestDecodeJSON_BasePass(t *testing.T) {
	binding := decodeBaseBinding()
	requireDecodePass(t, decodeDocument(nil, nil), binding)
	check(t, len(bindingFrame(binding)) == 772, "binding frame length")
}
func TestDecodeJSON_ParserReasons(t *testing.T) {
	requireDecodeFault(t, []byte{0xEF, 0xBB, 0xBF, '{', '}', 0xFF}, ReasonBOMForbidden)
	requireDecodeFault(t, []byte{'{', '}', 0xFF}, ReasonInvalidUTF8)
	requireDecodeFault(t, []byte("{"), ReasonInvalidJSON)
	requireDecodeFault(t, []byte("{}null"), ReasonTrailingValue)
	requireDecodeFault(t, []byte(`{"a":1,"a":2}`), ReasonDuplicateKey)
	requireDecodeFault(t, []byte("null"), ReasonInvalidSchema)
}
func TestDecodeJSON_EnvelopeReasons(t *testing.T) {
	requireDecodeFault(t, []byte("{}"), ReasonRequiredInputMissing)
	requireDecodeFault(t, []byte(`{"expected_label":"ignored"}`), ReasonUnknownField)
	requireDecodeFault(t, []byte(`{"schema":null}`), ReasonNullValue)
	requireDecodeFault(t, []byte(`{"schema":""}`), ReasonEmptyValue)
	requireDecodeFault(t, []byte(`{"schema":1}`), ReasonInvalidSchema)
	requireDecodeFault(t, []byte(`{"schema":"gooo/safe-work-binding/v1","task_id":"billing://"}`),
		ReasonInvalidStableID)
	requireDecodeFault(t, decodeOverride("source_snapshot_digest", `"invalid"`), ReasonInvalidDigest)
	requireDecodeFault(t, decodeOverride("binding_digest", `"sha256:`+strings.Repeat("0", 64)+`"`),
		ReasonBindingDigestMismatch)
}
func TestDecodeJSON_MixedPrecedence(t *testing.T) {
	requireDecodeFault(t, []byte(`{"expected":"ignored"}`), ReasonUnknownField)
	requireDecodeFault(t, []byte(`{"task_id":null}`), ReasonRequiredInputMissing)
	requireDecodeFault(t, []byte(`{"schema":null,"task_id":1}`), ReasonNullValue)
	requireDecodeFault(t, []byte(`{"schema":1,"task_id":null}`), ReasonInvalidSchema)
	requireDecodeFault(t, []byte(`{"a":1,"a":2`), ReasonInvalidJSON)
	requireDecodeFault(t, []byte(`{"task_id":[{"x":1,"x":2}]}`), ReasonDuplicateKey)
}
func TestDecodeJSON_StringTokens(t *testing.T) {
	requireDecodeFault(t, []byte(`{"schema":"\uFFFD"}`), ReasonInvalidSchema)
	requireDecodeFault(t, []byte(`{"schema":"\q"}`), ReasonInvalidJSON)
	requireDecodeFault(t, []byte(`{"schema":"\uD800"}`), ReasonInvalidJSON)
}
