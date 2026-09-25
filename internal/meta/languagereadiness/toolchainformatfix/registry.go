package toolchainformatfix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func DecodeRegistry(raw []byte) (Registry, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	registry := Registry{}
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Registry{}, fmt.Errorf("toolchain format/fix registry has trailing JSON")
	}
	if !reflect.DeepEqual(registry, expectedRegistry()) {
		return Registry{}, fmt.Errorf("toolchain format/fix registry contract mismatch")
	}
	return registry, nil
}

func expectedRegistry() Registry {
	return Registry{Schema: RegistrySchema, Version: RegistryVersion, Cases: []Definition{
		{ID: "positive-format-text", Kind: "POSITIVE", Operation: "FORMAT_TEXT", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "emit-canonical-source"},
		{ID: "positive-format-json", Kind: "POSITIVE", Operation: "FORMAT_JSON", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "emit-structured-format"},
		{ID: "positive-format-check", Kind: "POSITIVE", Operation: "FORMAT_CHECK", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "recognize-format-fixed-point"},
		{ID: "positive-fix-json-change", Kind: "POSITIVE", Operation: "FIX_JSON_CHANGED", ExpectedExit: 0, ProofChoice: "FOUNDATION", MetaOperation: "emit-versioned-fix-plan"},
		{ID: "positive-fix-json-fixed", Kind: "POSITIVE", Operation: "FIX_JSON_FIXED", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "emit-explicit-fixed-point"},
		{ID: "positive-fix-text", Kind: "POSITIVE", Operation: "FIX_TEXT", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "render-fix-plan"},
		{ID: "guardrail-format-required", Kind: "GUARDRAIL", Operation: "FORMAT_REQUIRED", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-missing-format-input"},
		{ID: "guardrail-format-malformed", Kind: "GUARDRAIL", Operation: "FORMAT_MALFORMED", ExpectedExit: 1, ProofChoice: "REGRESSION", MetaOperation: "lower-malformed-format-resolution"},
		{ID: "guardrail-fix-malformed", Kind: "GUARDRAIL", Operation: "FIX_MALFORMED", ExpectedExit: 1, ProofChoice: "REGRESSION", MetaOperation: "lower-malformed-fix-resolution"},
		{ID: "guardrail-format-usage", Kind: "GUARDRAIL", Operation: "FORMAT_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-unknown-format-flag"},
		{ID: "guardrail-fix-usage", Kind: "GUARDRAIL", Operation: "FIX_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-unknown-fix-flag"},
		{ID: "guardrail-fix-write", Kind: "GUARDRAIL", Operation: "FIX_FLAG", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-direct-write-authority"},
	}}
}
