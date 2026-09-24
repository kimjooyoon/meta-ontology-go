package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func loadRuntimePlanContract(reader SourceReader, filename string, source []byte, plan valueexecution.Plan) (string, error) {
	raw, err := reader.ReadFile(filename)
	if err != nil {
		return "", valueexecution.Failure{
			Code: valueexecution.ReasonSourceReadFailed, Stage: "PLAN", Step: "read-runtime-plan", Detail: err.Error(),
		}
	}
	digest, err := validateRuntimePlanContract(source, plan, raw)
	if err != nil {
		return "", valueexecution.Failure{
			Code: valueexecution.ReasonPlanInvalid, Stage: "PLAN", Step: "validate-runtime-plan-contract", Detail: err.Error(),
		}
	}
	return digest, nil
}

func validateRuntimePlanContract(source []byte, plan valueexecution.Plan, raw []byte) (string, error) {
	var document runtimePlanDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return "", fmt.Errorf("decode runtime plan: %w", err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return "", fmt.Errorf("decode runtime plan: %w", err)
	}
	if document.Schema != runtimePlanSchema {
		return "", fmt.Errorf("unsupported runtime plan schema %q", document.Schema)
	}
	expectedSourceDigest := runtimePlanSourceDigest(source)
	if document.SourceDigest != expectedSourceDigest || plan.SourceDigest != expectedSourceDigest {
		return "", fmt.Errorf("source digest mismatch: artifact=%q expected=%q plan=%q", document.SourceDigest, expectedSourceDigest, plan.SourceDigest)
	}
	if document.RuntimeBindingCount != plan.RuntimeBindingCount() {
		return "", fmt.Errorf("runtime binding count mismatch: artifact=%d expected=%d", document.RuntimeBindingCount, plan.RuntimeBindingCount())
	}

	file, diagnostics := syntax.ParseFile("runtime-plan-contract.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return "", fmt.Errorf("source parse failed while validating runtime plan")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return "", fmt.Errorf("lower source for runtime plan identity: %w", err)
	}
	if document.SemanticHash != ir.StableHash() {
		return "", fmt.Errorf("semantic hash mismatch: artifact=%q expected=%q", document.SemanticHash, ir.StableHash())
	}

	documentModel, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil {
		return "", fmt.Errorf("build typed-plan document: %w", err)
	}
	if len(documentModel.BindingEdges) > 0 {
		typedPlan, err := bidir.CompileTypedPlan(documentModel)
		if err != nil {
			return "", fmt.Errorf("compile typed plan: %w", err)
		}
		if document.TypedPlanDigest == "" {
			return "", fmt.Errorf("runtime plan omits typed-plan identity")
		}
		if document.TypedPlanDigest != typedPlan.Digest() {
			return "", fmt.Errorf("typed-plan digest mismatch: artifact=%q expected=%q", document.TypedPlanDigest, typedPlan.Digest())
		}
		if len(document.ActivityOrder) != len(typedPlan.Activities) {
			return "", fmt.Errorf("typed-plan activity order length mismatch: artifact=%d expected=%d", len(document.ActivityOrder), len(typedPlan.Activities))
		}
		for index, activity := range typedPlan.Activities {
			if document.ActivityOrder[index] != string(activity) {
				return "", fmt.Errorf("typed-plan activity order mismatch at %d: artifact=%q expected=%q", index, document.ActivityOrder[index], activity)
			}
		}
		if len(document.BindingEdgeOrder) != len(typedPlan.Edges) {
			return "", fmt.Errorf("typed-plan binding edge order length mismatch: artifact=%d expected=%d", len(document.BindingEdgeOrder), len(typedPlan.Edges))
		}
		for index, edge := range typedPlan.Edges {
			want := runtimePlanBindingEdgeKey(edge)
			if document.BindingEdgeOrder[index] != want {
				return "", fmt.Errorf("typed-plan binding edge order mismatch at %d: artifact=%q expected=%q", index, document.BindingEdgeOrder[index], want)
			}
		}
	} else if document.TypedPlanDigest != "" || len(document.ActivityOrder) != 0 || len(document.BindingEdgeOrder) != 0 {
		return "", fmt.Errorf("runtime plan contains typed-plan identity without explicit binding edges")
	}

	return "sha256:" + cache.HashBytes(raw).String(), nil
}
