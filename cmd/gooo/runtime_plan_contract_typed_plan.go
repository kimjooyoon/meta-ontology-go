package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func validateRuntimePlanTypedPlan(file *syntax.File, document runtimePlanDocument) error {
	documentModel, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil {
		return fmt.Errorf("build typed-plan document: %w", err)
	}
	if len(documentModel.BindingEdges) > 0 {
		typedPlan, err := bidir.CompileTypedPlan(documentModel)
		if err != nil {
			return fmt.Errorf("compile typed plan: %w", err)
		}
		if document.TypedPlanDigest == "" {
			return fmt.Errorf("runtime plan omits typed-plan identity")
		}
		if document.TypedPlanDigest != typedPlan.Digest() {
			return fmt.Errorf("typed-plan digest mismatch: artifact=%q expected=%q", document.TypedPlanDigest, typedPlan.Digest())
		}
		if len(document.ActivityOrder) != len(typedPlan.Activities) {
			return fmt.Errorf("typed-plan activity order length mismatch: artifact=%d expected=%d", len(document.ActivityOrder), len(typedPlan.Activities))
		}
		for index, activity := range typedPlan.Activities {
			if document.ActivityOrder[index] != string(activity) {
				return fmt.Errorf("typed-plan activity order mismatch at %d: artifact=%q expected=%q", index, document.ActivityOrder[index], activity)
			}
		}
		if len(document.BindingEdgeOrder) != len(typedPlan.Edges) {
			return fmt.Errorf("typed-plan binding edge order length mismatch: artifact=%d expected=%d", len(document.BindingEdgeOrder), len(typedPlan.Edges))
		}
		for index, edge := range typedPlan.Edges {
			want := runtimePlanBindingEdgeKey(edge)
			if document.BindingEdgeOrder[index] != want {
				return fmt.Errorf("typed-plan binding edge order mismatch at %d: artifact=%q expected=%q", index, document.BindingEdgeOrder[index], want)
			}
		}
	} else if document.TypedPlanDigest != "" || len(document.ActivityOrder) != 0 || len(document.BindingEdgeOrder) != 0 {
		return fmt.Errorf("runtime plan contains typed-plan identity without explicit binding edges")
	}
	return nil
}
