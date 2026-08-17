package safeworkbinding

import (
	"reflect"
	"strings"
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
	requireEnvelopeReason(t, baseEnvelopeValue(), ReasonNone)
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

func TestValidateEnvelope_FixedFieldOrder(t *testing.T) {
	empty := jsonValue{kind: jsonObjectValue, object: map[string]jsonValue{}}
	requireEnvelopeReason(t, empty, ReasonRequiredInputMissing)
	value := baseEnvelopeValue()
	value.object["schema"] = envelopeString("invalid")
	delete(value.object, "task_id")
	requireEnvelopeReason(t, value, ReasonInvalidSchema)
	value = baseEnvelopeValue()
	delete(value.object, "schema")
	value.object["task_id"] = jsonValue{kind: jsonNullValue}
	requireEnvelopeReason(t, value, ReasonRequiredInputMissing)
	value = baseEnvelopeValue()
	value.object["schema"] = jsonValue{kind: jsonNullValue}
	value.object["task_id"] = jsonValue{kind: jsonNumberValue}
	requireEnvelopeReason(t, value, ReasonNullValue)
	value = baseEnvelopeValue()
	value.object["schema"] = jsonValue{kind: jsonNumberValue}
	value.object["task_id"] = jsonValue{kind: jsonNullValue}
	requireEnvelopeReason(t, value, ReasonInvalidSchema)
	value = baseEnvelopeValue()
	value.object["task_id"] = envelopeString("billing://")
	delete(value.object, "path_id")
	requireEnvelopeReason(t, value, ReasonInvalidStableID)
	value = baseEnvelopeValue()
	value.object["source_snapshot_digest"] = envelopeString("invalid")
	delete(value.object, "semantic_snapshot_digest")
	requireEnvelopeReason(t, value, ReasonInvalidDigest)
}

func TestValidateEnvelope_PresenceAndTypes(t *testing.T) {
	requireEnvelopeReason(t, envelopeString("wrong root"), ReasonInvalidSchema)
	requireEnvelopeReason(t, jsonValue{kind: jsonObjectValue}, ReasonInvalidSchema)
	for _, field := range bindingFieldOrder {
		value := baseEnvelopeValue()
		delete(value.object, field)
		requireEnvelopeReason(t, value, ReasonRequiredInputMissing)
	}
	requireEnvelopeReason(t, envelopeWithField("schema", jsonValue{kind: jsonNullValue}), ReasonNullValue)
	requireEnvelopeReason(t, envelopeWithField("schema", envelopeString("")), ReasonEmptyValue)
	requireEnvelopeReason(t, envelopeWithField("schema", jsonValue{kind: jsonNumberValue}), ReasonInvalidSchema)
	requireEnvelopeReason(t, envelopeWithField("schema", jsonValue{kind: jsonBoolValue}), ReasonInvalidSchema)
	requireEnvelopeReason(t, envelopeWithField("schema", jsonValue{kind: jsonArrayValue}), ReasonInvalidSchema)
	object := jsonValue{kind: jsonObjectValue, object: map[string]jsonValue{}}
	requireEnvelopeReason(t, envelopeWithField("schema", object), ReasonInvalidSchema)
}

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
	if !validateDigest("sha256:" + strings.Repeat("0", 64)) {
		t.Fatal("valid digest rejected")
	}
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
