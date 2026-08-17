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

func checkResultShape(t *testing.T, result ParseResult, want Reason) {
	t.Helper()
	check(t, result.Reason == want && result.ResultDigest != "", "result reason and digest")
	check(t, !result.ExecutionAuthorized && result.EnforcementEffect == EnforcementEffectNoEffect, "result safety")
	if want == ReasonNone {
		check(t, result.Decision == DecisionPass && !result.FullSuiteRequired, "pass state")
		check(t, result.Faults != nil && len(result.Faults) == 0, "pass faults")
		return
	}
	decision := DecisionFailClosed
	fullSuite := false
	if want == ReasonRequiredInputMissing {
		decision = DecisionUnknown
		fullSuite = true
	}
	check(t, result.Decision == decision && result.FullSuiteRequired == fullSuite, "fault state")
	check(t, result.Faults != nil && len(result.Faults) == 1 && result.Faults[0] == want, "faults")
	check(t, result.ReplayDigest == "", "fault replay")
}

func requireDecodeFault(t *testing.T, input []byte, want Reason) ParseResult {
	t.Helper()
	binding, result := DecodeJSON(input)
	check(t, binding == (SafeWorkBinding{}), "fault binding")
	checkResultShape(t, result, want)
	return result
}

func requireDecodePass(t *testing.T, input []byte, want SafeWorkBinding) ParseResult {
	t.Helper()
	binding, result := DecodeJSON(input)
	check(t, binding == want, "pass binding")
	checkResultShape(t, result, ReasonNone)
	check(t, result.ResultDigest == digestPass, "pass result digest")
	check(t, result.ReplayDigest == replayDigest(want.BindingDigest, digestPass), "pass replay")
	return result
}

func checkGovernedDecode(t *testing.T, field, value string) {
	t.Helper()
	binding := decodeBaseBinding()
	mutateEnvelopeBinding(&binding, field, value)
	overrides := map[string]string{field: `"` + value + `"`}
	requireDecodeFault(t, decodeDocument(nil, overrides), ReasonBindingDigestMismatch)
	binding.BindingDigest = bindingDigest(binding)
	overrides["binding_digest"] = `"` + string(binding.BindingDigest) + `"`
	requireDecodePass(t, decodeDocument(nil, overrides), binding)
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

func TestDecodeJSON_UnknownIsolation(t *testing.T) {
	fields := []string{
		"expected",
		"expected_label",
		"want",
		"legacy_work_id",
		"result_digest",
		"replay_digest",
	}
	values := []string{
		`"ignored"`,
		"null",
		"{}",
	}
	for _, field := range fields {
		for _, value := range values {
			result := requireDecodeFault(t, []byte(`{"`+field+`":`+value+`}`), ReasonUnknownField)
			check(t, result.ResultDigest == digestExpectedLabel, "unknown result isolation")
		}
	}
}

func TestDecodeJSON_StringTokens(t *testing.T) {
	requireDecodeFault(t, []byte(`{"schema":"\uFFFD"}`), ReasonInvalidSchema)
	requireDecodeFault(t, []byte(`{"schema":"\q"}`), ReasonInvalidJSON)
	requireDecodeFault(t, []byte(`{"schema":"\uD800"}`), ReasonInvalidJSON)
}

func TestDecodeJSON_StableIDAndDigestGrammar(t *testing.T) {
	boundary := "billing://entity/" + strings.Repeat("a", 239)
	checkGovernedDecode(t, "task_id", boundary)
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
	binding := decodeBaseBinding()
	binding.Schema = "invalid"
	binding.BindingDigest = bindingDigest(binding)
	requireDecodeFault(t, decodeDocument(nil, map[string]string{
		"schema": `"invalid"`, "binding_digest": `"` + string(binding.BindingDigest) + `"`,
	}), ReasonInvalidSchema)
	checkGovernedDecode(t, "task_id", "billing://task/pay-v2")
	checkGovernedDecode(t, "path_id", "billing://path/pay-v2")
	checkGovernedDecode(t, "obligation_id", "billing://obligation/pay-v2")
	checkGovernedDecode(t, "source_snapshot_digest", "sha256:"+strings.Repeat("6", 64))
	checkGovernedDecode(t, "semantic_snapshot_digest", "sha256:"+strings.Repeat("7", 64))
	checkGovernedDecode(t, "policy_digest", "sha256:"+strings.Repeat("8", 64))
	checkGovernedDecode(t, "registry_digest", "sha256:"+strings.Repeat("9", 64))
	checkGovernedDecode(t, "toolchain_options_digest", "sha256:"+strings.Repeat("a", 64))
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

func TestDecodeJSON_ResultReplayVectors(t *testing.T) {
	vectors := []struct {
		reason Reason
		length int
		digest Digest
	}{
		{ReasonNone, 255, digestPass},
		{ReasonRequiredInputMissing, 306, digestMissing},
		{ReasonInvalidUTF8, 290, "sha256:cb0f91bbeb15fdabbf323f49861cf155dceaba0e58eb3b7f02ffd75d6152e3cf"},
		{ReasonBOMForbidden, 292, digestBOMMixed},
		{ReasonInvalidJSON, 290, "sha256:6ffcacd5fca7d1ad2b4449bdb2777f3cc70af29d4cc85be8677496faeb0fdb66"},
		{ReasonTrailingValue, 294, "sha256:bda8407767c9c12ac9bd5a5adc2723a436a01f63e748fd68f15f843a5c4afb4b"},
		{ReasonDuplicateKey, 292, digestDuplicate},
		{ReasonUnknownField, 292, digestExpectedLabel},
		{ReasonNullValue, 286, "sha256:d125c223de622549e7216bf546d817c630d9e6a89e863343598718b1b0117287"},
		{ReasonEmptyValue, 288, "sha256:5e9b99802e1a4e8eee8035b60f63bc0e2b8a34d0a33df108be67e282c43dcfd4"},
		{ReasonInvalidSchema, 294, "sha256:46bac36b14f9e6da66d01fefef6d62300f9fbc0e28e46c906f7661e3569a24fb"},
		{ReasonInvalidStableID, 300, "sha256:9650730414b4329361cb07e48c21ecf3a6e56e4fc400f284ccf2b5a0b8b20873"},
		{ReasonInvalidDigest, 294, "sha256:4cc1918712c3f741c7fb522c13cbc773a97343b8c675d5b1da90ecd9a5b8c944"},
		{ReasonBindingDigestMismatch, 312, digestMismatch},
	}
	for _, tc := range vectors {
		result := resultForPass()
		if tc.reason != ReasonNone {
			result = resultForReason(tc.reason)
		}
		result = completeResult(result)
		frame, ok := resultFrame(result)
		check(t, ok && len(frame) == tc.length, "result frame vector")
		check(t, result.ResultDigest == tc.digest, "result digest vector")
		checkResultShape(t, result, tc.reason)
		if tc.reason == ReasonNone {
			check(t, len(replayFrame(digestBase, result.ResultDigest)) == 252, "replay frame length")
			check(t, replayDigest(digestBase, result.ResultDigest) == digestReplay, "replay digest vector")
		}
	}
}
