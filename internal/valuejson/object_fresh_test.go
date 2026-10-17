package valuejson

import (
	"strings"
	"testing"
)

func TestSharedObjectLexicalAdmissionFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name, raw, reason string
	}{
		{"duplicate", `{"x":1,"x":2}`, "distinct names"},
		{"key-surrogate", `{"\ud800":1,"�":2}`, "UNPAIRED_HIGH"},
		{"value-surrogate", `{"x":"\udfff"}`, "UNPAIRED_LOW"},
		{"valid-pair", `{"x":"\ud83e\udded"}`, ""},
		{"explicit-replacement", `{"x":"\ufffd"}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values, err := Object([]byte(tc.raw), 32)
			if tc.reason == "" {
				if err != nil || len(values) != 1 {
					t.Fatal(values, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.reason) || values != nil {
				t.Fatal("lexical source ambiguity normalized", values, err)
			}
		})
	}
}
