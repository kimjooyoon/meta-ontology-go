package assemblyspec

import "testing"

func TestTypedValueJSONKeepsExactIntegersAndDistinctKeys(t *testing.T) {
	value, err := CanonicalValue(`[{"z":9223372036854775807,"a":-9223372036854775808},true,"한글"]`)
	if err != nil || value != `[{"a":-9223372036854775808,"z":9223372036854775807},true,"한글"]` {
		t.Fatal(value, err)
	}
	for _, raw := range []string{`{"a":1,"a":2}`, `[{"a":1,"\u0061":2}]`, `{"a":{"b":1,"b":2}}`, `[] []`, `{"a":1`, `null true`} {
		if _, err := CanonicalValue(raw); err == nil {
			t.Fatal("ambiguous value accepted", raw)
		}
	}
}
