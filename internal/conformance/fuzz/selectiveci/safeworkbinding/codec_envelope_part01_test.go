package safeworkbinding

import (
	"reflect"
	"testing"
)

func envelopeString(value string) jsonValue {
	return jsonValue{kind: jsonStringValue, text: value}
}
func baseEnvelopeValue() jsonValue {
	binding := reflect.ValueOf(baseBindingForDigest())
	object := map[string]jsonValue{"binding_digest": envelopeString(digestBase)}
	for index, field := range bindingFieldOrder[:9] {
		object[field] = envelopeString(binding.Field(index).String())
	}
	return jsonValue{kind: jsonObjectValue, object: object}
}
func envelopeWithField(field string, fieldValue jsonValue) jsonValue {
	value := baseEnvelopeValue()
	value.object[field] = fieldValue
	return value
}
func requireEnvelopeReason(t *testing.T, value jsonValue, want Reason) SafeWorkBinding {
	t.Helper()
	binding, reason := validateEnvelope(value)
	if reason != want {
		t.Fatalf("reason=%v, want %v", reason, want)
	}
	if reason != ReasonNone && binding != (SafeWorkBinding{}) {
		t.Errorf("failure returned binding %#v", binding)
	}
	return binding
}
func mutateEnvelopeBinding(binding *SafeWorkBinding, field, value string) {
	reflect.ValueOf(map[string]any{
		"task_id":                  &binding.TaskID,
		"path_id":                  &binding.PathID,
		"obligation_id":            &binding.ObligationID,
		"source_snapshot_digest":   &binding.SourceSnapshotDigest,
		"semantic_snapshot_digest": &binding.SemanticSnapshotDigest,
		"policy_digest":            &binding.PolicyDigest,
		"registry_digest":          &binding.RegistryDigest,
		"toolchain_options_digest": &binding.ToolchainOptionsDigest,
	}[field]).Elem().SetString(value)
}
func TestValidateEnvelope_Base(t *testing.T) {
	want := baseBindingForDigest()
	want.BindingDigest = digestBase
	check(t, requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone) == want, "base binding mismatch")
}
func TestValidateEnvelope_UnknownBeforeMissing(t *testing.T) {
	for _, field := range []string{
		"expected",
		"expected_label",
		"want",
		"legacy_work_id",
		"result_digest",
		"replay_digest",
	} {
		requireEnvelopeReason(t, envelopeWithField(field, envelopeString("ignored")), ReasonUnknownField)
	}
	for _, extra := range []jsonValue{
		{kind: jsonNullValue},
		{kind: jsonObjectValue, object: map[string]jsonValue{}},
	} {
		requireEnvelopeReason(t, envelopeWithField("expected_label", extra), ReasonUnknownField)
	}
	unknownOnly := jsonValue{kind: jsonObjectValue, object: map[string]jsonValue{"expected": envelopeString("ignored")}}
	requireEnvelopeReason(t, unknownOnly, ReasonUnknownField)
}
