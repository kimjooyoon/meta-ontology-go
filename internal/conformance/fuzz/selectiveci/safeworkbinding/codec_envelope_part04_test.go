package safeworkbinding

import (
	"strings"
	"testing"
)

func TestValidateEnvelope_ConstructionOrder(t *testing.T) {
	source := baseEnvelopeValue()
	want := requireEnvelopeReason(t, source, ReasonNone)
	reversed := jsonValue{kind: jsonObjectValue, object: make(map[string]jsonValue)}
	for index := len(bindingFieldOrder) - 1; index >= 0; index-- {
		field := bindingFieldOrder[index]
		reversed.object[field] = source.object[field]
	}
	if got := requireEnvelopeReason(t, reversed, ReasonNone); got != want {
		t.Fatalf("reversed binding=%#v", got)
	}
}
func TestValidateEnvelope_WhitespaceLocation(t *testing.T) {
	input := []byte("\n{\n" +
		"  \"schema\" : \"gooo/safe-work-binding/v1\",\n" +
		"  \"task_id\" : \"billing://task/pay\",\n" +
		"  \"path_id\" : \"billing://path/pay\",\n" +
		"  \"obligation_id\" : \"billing://obligation/pay\",\n" +
		"  \"source_snapshot_digest\" : \"sha256:" + strings.Repeat("1", 64) + "\",\n" +
		"  \"semantic_snapshot_digest\" : \"sha256:" + strings.Repeat("2", 64) + "\",\n" +
		"  \"policy_digest\" : \"sha256:" + strings.Repeat("3", 64) + "\",\n" +
		"  \"registry_digest\" : \"sha256:" + strings.Repeat("4", 64) + "\",\n" +
		"  \"toolchain_options_digest\" : \"sha256:" + strings.Repeat("5", 64) + "\",\n" +
		"  \"binding_digest\" : \"sha256:dc6dbe157ede5924b61676bfdcd4151cd6f73a51b7eefda674cca3d6d169a5cb\"\n}\n")
	document := requireDocumentReason(t, input, ReasonNone)
	want := requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone)
	if got := requireEnvelopeReason(t, document, ReasonNone); got != want {
		t.Fatalf("whitespace binding=%#v", got)
	}
}
func TestValidateEnvelope_GovernedMutations(t *testing.T) {
	binding := requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone)
	binding.Schema = "invalid"
	value := envelopeWithField("schema", envelopeString(binding.Schema))
	value.object["binding_digest"] = envelopeString(string(bindingDigest(binding)))
	requireEnvelopeReason(t, value, ReasonInvalidSchema)
	cases := []struct {
		field string
		value string
	}{
		{field: "task_id", value: "billing://task/pay-v2"},
		{field: "path_id", value: "billing://path/pay-v2"},
		{field: "obligation_id", value: "billing://obligation/pay-v2"},
		{field: "source_snapshot_digest", value: "sha256:" + strings.Repeat("6", 64)},
		{field: "semantic_snapshot_digest", value: "sha256:" + strings.Repeat("7", 64)},
		{field: "policy_digest", value: "sha256:" + strings.Repeat("8", 64)},
		{field: "registry_digest", value: "sha256:" + strings.Repeat("9", 64)},
		{field: "toolchain_options_digest", value: "sha256:" + strings.Repeat("a", 64)},
	}
	for _, tc := range cases {
		t.Run(tc.field, func(t *testing.T) {
			value := envelopeWithField(tc.field, envelopeString(tc.value))
			requireEnvelopeReason(t, value, ReasonBindingDigestMismatch)
			binding := requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone)
			mutateEnvelopeBinding(&binding, tc.field, tc.value)
			binding.BindingDigest = bindingDigest(binding)
			value.object["binding_digest"] = envelopeString(string(binding.BindingDigest))
			if got := requireEnvelopeReason(t, value, ReasonNone); got != binding {
				t.Fatalf("binding=%#v, want %#v", got, binding)
			}
		})
	}
}
