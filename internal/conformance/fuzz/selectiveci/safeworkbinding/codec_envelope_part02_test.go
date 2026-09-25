package safeworkbinding

import (
	"testing"
)

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
