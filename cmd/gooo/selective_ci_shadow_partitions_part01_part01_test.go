package main

import (
	"bytes"
	"testing"
)

func TestSelectiveCIShadowMalformedUnknownAndDuplicateJSONFallback(t *testing.T) {
	cases := []struct {
		name      string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "malformed", component: "evidence_input", mutate: func(f *shadowFixture) { f.files["evidence_input.json"] = []byte("{") }},
		{name: "unknown field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = append(f.files["evidence_input.json"], []byte(" ")...)
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("{"), []byte("{\"unknown\":true,"), 1)
		}},
		{name: "duplicate field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("\"schema\":"), []byte("\"schema\":\"gooo-selective-ci-evidence/v1\",\"schema\":"), 1)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != "INPUT" || output.Component != test.component || output.Reason == "" {
				t.Fatalf("fallback = %#v", output)
			}
		})
	}
}
