package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func decodeRuntimePlanContract(raw []byte) (runtimePlanDocument, error) {
	var document runtimePlanDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return runtimePlanDocument{}, fmt.Errorf("decode runtime plan: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return runtimePlanDocument{}, fmt.Errorf("decode runtime plan: %w", err)
	}
	if document.Schema != runtimePlanSchema {
		return runtimePlanDocument{}, fmt.Errorf("unsupported runtime plan schema %q", document.Schema)
	}
	return document, nil
}

func validateRuntimePlanIdentity(source []byte, plan valueexecution.Plan, document runtimePlanDocument) (*syntax.File, error) {
	expectedSourceDigest := runtimePlanSourceDigest(source)
	if document.SourceDigest != expectedSourceDigest || plan.SourceDigest != expectedSourceDigest {
		return nil, fmt.Errorf("source digest mismatch: artifact=%q expected=%q plan=%q", document.SourceDigest, expectedSourceDigest, plan.SourceDigest)
	}
	if document.RuntimeBindingCount != plan.RuntimeBindingCount() {
		return nil, fmt.Errorf("runtime binding count mismatch: artifact=%d expected=%d", document.RuntimeBindingCount, plan.RuntimeBindingCount())
	}
	file, diagnostics := syntax.ParseFile("runtime-plan-contract.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return nil, fmt.Errorf("source parse failed while validating runtime plan")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return nil, fmt.Errorf("lower source for runtime plan identity: %w", err)
	}
	if document.SemanticHash != ir.StableHash() {
		return nil, fmt.Errorf("semantic hash mismatch: artifact=%q expected=%q", document.SemanticHash, ir.StableHash())
	}
	if err := validateRuntimePlanBindings(ir, document); err != nil {
		return nil, err
	}
	return file, nil
}

func validateRuntimePlanBindings(ir semantic.IR, document runtimePlanDocument) error {
	expected := runtimePlanBindings(ir)
	if len(document.Bindings) != len(expected) {
		return fmt.Errorf("runtime binding table length mismatch: artifact=%d expected=%d", len(document.Bindings), len(expected))
	}
	for index, binding := range expected {
		if document.Bindings[index] != binding {
			return fmt.Errorf("runtime binding mismatch at %d: artifact=%+v expected=%+v", index, document.Bindings[index], binding)
		}
	}
	return nil
}
