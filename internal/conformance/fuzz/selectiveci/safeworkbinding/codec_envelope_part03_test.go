package safeworkbinding

import (
	"strings"
	"testing"
)

func TestValidateEnvelope_StableIDs(t *testing.T) {
	prefix := "billing://entity/"
	boundary := prefix + strings.Repeat("a", 239)
	tooLong := prefix + strings.Repeat("a", 240)
	cases := []struct {
		value string
		valid bool
	}{
		{value: "billing://entity/Order", valid: true},
		{value: "urn:gooo:entity:Order", valid: true},
		{value: " billing://entity/Order", valid: false},
		{value: "billing://entity/Order ", valid: false},
		{value: "Billing://Entity/Order", valid: false},
		{value: "billing://Entity/Order", valid: false},
		{value: "billing://", valid: false},
		{value: "entity/Order", valid: false},
		{value: "billing://entity/Order Name", valid: false},
		{value: "billing://entity/Order\x01", valid: false},
		{value: boundary, valid: true},
		{value: tooLong, valid: false},
		{value: string([]byte{0xFF}), valid: false},
	}
	for _, tc := range cases {
		if validateStableID(tc.value) != tc.valid {
			t.Errorf("validateStableID(%q)=%v, want %v", tc.value, !tc.valid, tc.valid)
		}
	}
	if len(boundary) != 256 {
		t.Fatalf("boundary length=%d, want 256", len(boundary))
	}
	binding := requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone)
	binding.TaskID = StableID(boundary)
	value := envelopeWithField("task_id", envelopeString(boundary))
	value.object["binding_digest"] = envelopeString(string(bindingDigest(binding)))
	requireEnvelopeReason(t, value, ReasonNone)
	value = envelopeWithField("task_id", envelopeString(tooLong))
	requireEnvelopeReason(t, value, ReasonInvalidStableID)
}
func TestValidateEnvelope_Digests(t *testing.T) {
	check(t, validateDigest("sha256:"+strings.Repeat("0", 64)), "valid digest rejected")
	for _, value := range []string{
		"sha256:" + strings.Repeat("A", 64),
		"SHA256:" + strings.Repeat("0", 64),
		strings.Repeat("0", 64),
		"sha256:" + strings.Repeat("0", 63),
		"sha256:" + strings.Repeat("0", 65),
		"sha256:" + strings.Repeat("0", 63) + "g",
	} {
		if validateDigest(value) {
			t.Errorf("invalid digest accepted: %q", value)
		}
	}
}
func TestValidateEnvelope_StaleBinding(t *testing.T) {
	value := envelopeWithField("binding_digest", envelopeString("sha256:"+strings.Repeat("0", 64)))
	requireEnvelopeReason(t, value, ReasonBindingDigestMismatch)
}
