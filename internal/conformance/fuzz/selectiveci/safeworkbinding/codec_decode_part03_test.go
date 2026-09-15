package safeworkbinding

import (
	"strings"
	"testing"
)

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
