package safeworkbinding

import (
	"strings"
	"testing"
)

func decodeBaseBinding() SafeWorkBinding {
	binding := baseBindingForDigest()
	binding.BindingDigest = digestBase
	return binding
}

func decodeValues() map[string]string {
	values := make(map[string]string, len(bindingFieldOrder))
	for field, value := range baseEnvelopeValue().object {
		values[field] = `"` + value.text + `"`
	}
	return values
}

func decodeDocument(order []string, overrides map[string]string) []byte {
	values := decodeValues()
	for field, value := range overrides {
		values[field] = value
	}
	if order == nil {
		order = bindingFieldOrder[:]
	}
	var document strings.Builder
	document.WriteByte('{')
	for index, field := range order {
		if index > 0 {
			document.WriteByte(',')
		}
		document.WriteString(`"` + field + `":` + values[field])
	}
	document.WriteByte('}')
	return []byte(document.String())
}

func decodeOverride(field, value string) []byte {
	return decodeDocument(nil, map[string]string{field: value})
}

func whitespaceDecodeDocument() []byte {
	document := string(decodeDocument(nil, nil))
	document = strings.Replace(document, `{"`, "\n{\n  \"", 1)
	document = strings.ReplaceAll(document, `":"`, `" : "`)
	document = strings.ReplaceAll(document, `","`, "\",\n  \"")
	document = strings.Replace(document, `"}`, "\"\n}\n", 1)
	return []byte(document)
}

func requireDecodeFault(t *testing.T, input []byte, want Reason) ParseResult {
	t.Helper()
	binding, result := DecodeJSON(input)
	check(t, binding == (SafeWorkBinding{}), "fault binding")
	check(t, result.Reason == want, "fault reason")
	check(t, result.Faults != nil && len(result.Faults) == 1, "fault list")
	check(t, result.Faults[0] == want, "fault value")
	check(t, result.ResultDigest != "", "fault result digest")
	check(t, result.ReplayDigest == "", "fault replay")
	check(t, !result.ExecutionAuthorized, "fault authorization")
	check(t, result.EnforcementEffect == EnforcementEffectNoEffect, "fault effect")
	return result
}

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

func TestDecodeJSON_StableIDAndDigestGrammar(t *testing.T) {
	boundary := "billing://entity/" + strings.Repeat("a", 239)
	checkGovernedDecode(t, "task_id", boundary,
		"sha256:6f255ad87c0fc5d5c0a62b8f6dba6ec3516e13d2a614e3d673cda2d0ffb97579")
	tooLong := `{"schema":"gooo/safe-work-binding/v1","task_id":"billing://entity/` +
		strings.Repeat("a", 240) + `"}`
	requireDecodeFault(t, []byte(tooLong), ReasonInvalidStableID)
	requireDecodeFault(t, append(decodeDocument(nil, nil), 0xFF), ReasonInvalidUTF8)
	for _, value := range []string{
		"sha256:" + strings.Repeat("A", 64),
		"SHA256:" + strings.Repeat("0", 64),
		strings.Repeat("0", 64),
		"sha256:" + strings.Repeat("0", 63),
		"sha256:" + strings.Repeat("0", 65),
		"sha256:" + strings.Repeat("0", 63) + "g",
	} {
		requireDecodeFault(t, decodeOverride("source_snapshot_digest", `"`+value+`"`), ReasonInvalidDigest)
	}
}

func TestDecodeJSON_GovernedMutations(t *testing.T) {
	requireDecodeFault(t, decodeDocument(nil, map[string]string{
		"schema":         `"invalid"`,
		"binding_digest": `"sha256:20cb8d81572e1d4e06cadd039a4d89c8c394bea4dd2a128f769e0d961a789cab"`,
	}), ReasonInvalidSchema)
	checkGovernedDecode(t, "task_id", "billing://task/pay-v2",
		"sha256:5199c307c3e5ea18acf9d9299092c72b544d9c5273cf04232089c89b0c533441")
	checkGovernedDecode(t, "path_id", "billing://path/pay-v2",
		"sha256:c262c10f1652d0d3fd5fa605166d565a44261b834ffcffe876f2bb785f4bcd51")
	checkGovernedDecode(t, "obligation_id", "billing://obligation/pay-v2",
		"sha256:98735e6e37a893db3a2619f499e34bdb83086a5fc20c7ef70180c01a959b013b")
	checkGovernedDecode(t, "source_snapshot_digest", "sha256:"+strings.Repeat("6", 64),
		"sha256:a66b32b5c184ae1f879260bc0087765b67ae901a1b3850101df724d8b2f00596")
	checkGovernedDecode(t, "semantic_snapshot_digest", "sha256:"+strings.Repeat("7", 64),
		"sha256:24a448adbf4e668536178276a2c55d88ee948baec7150c1c4e5f4f3ba0a0db54")
	checkGovernedDecode(t, "policy_digest", "sha256:"+strings.Repeat("8", 64),
		"sha256:88f2275f7b8b4e9d2d759dec09f3c5bd1a00b607aa7dc7a1eb4f9b5ebb78e168")
	checkGovernedDecode(t, "registry_digest", "sha256:"+strings.Repeat("9", 64),
		"sha256:23477b01e3d477d6ba5702adca8ad77a2cd643748b29e20ac4a7fda0b020673e")
	checkGovernedDecode(t, "toolchain_options_digest", "sha256:"+strings.Repeat("a", 64),
		"sha256:758db32e3f69998d81177bca728f1b4bcfea90e8f9e3d5693d020eb5f5a1d843")
}

func TestDecodeJSON_MemberPermutation(t *testing.T) {
	binding := decodeBaseBinding()
	order := make([]string, len(bindingFieldOrder))
	for index := range bindingFieldOrder {
		order[index] = bindingFieldOrder[len(bindingFieldOrder)-1-index]
	}
	want := requireDecodePass(t, decodeDocument(nil, nil), binding)
	got := requireDecodePass(t, decodeDocument(order, nil), binding)
	check(t, got.ResultDigest == want.ResultDigest && got.ReplayDigest == want.ReplayDigest, "permutation result")
}

func TestDecodeJSON_RelocationWithValuesUnchanged(t *testing.T) {
	binding := decodeBaseBinding()
	want := requireDecodePass(t, decodeDocument(nil, nil), binding)
	got := requireDecodePass(t, whitespaceDecodeDocument(), binding)
	check(t, got.ResultDigest == want.ResultDigest && got.ReplayDigest == want.ReplayDigest, "whitespace result")
}
