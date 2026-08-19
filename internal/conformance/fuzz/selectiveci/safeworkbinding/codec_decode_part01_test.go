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
