package bodycodegen

import (
	"context"
	"encoding/json"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// CriteriaSourceFlow exposes the fixed source requirements beside the owner's
// alternatives. Existing model features remain separate from this source view.
type CriteriaSourceFlow struct {
	Schema              string                `json:"schema"`
	Status              string                `json:"status"`
	CandidatePlan       *pathplan.Plan        `json:"candidate_plan,omitempty"`
	CandidatePlanSHA256 string                `json:"candidate_plan_sha256,omitempty"`
	Predicates          []CriterionSourceFlow `json:"predicates"`
	Scope               string                `json:"scope"`
}

type CriterionSourceFlow struct {
	ID              string                          `json:"id"`
	Activity        string                          `json:"activity"`
	ActivityID      string                          `json:"activity_id"`
	GeneratedSHA256 string                          `json:"generated_sha256"`
	Inputs          []CriterionSourceInput          `json:"inputs"`
	ValueFlow       *RecordValueFlow                `json:"value_flow"`
	Nodes           []jointdecision.RecordGraphNode `json:"nodes,omitempty"`
	GraphSHA256     string                          `json:"graph_sha256,omitempty"`
}

type CriterionSourceInput struct {
	Input uint16 `json:"input"`
	Role  string `json:"role"`
	Type  string `json:"type"`
}

func criteriaSourceFlow(ctx context.Context, filename string, source []byte, owner recordAssemblyPlan) *CriteriaSourceFlow {
	if owner.spec.Criteria == nil {
		return nil
	}
	r := &CriteriaSourceFlow{Schema: "gooo/criteria-source-flow/v1", Status: "RESOLVED",
		Scope: "typed source alternatives and fixed Boolean predicate graphs; original input/result roles and helper provenance; finite inputs, expected outputs, held cases, question answers and model feature encoding excluded"}
	if owner.scalar != nil {
		plan := owner.scalar.sourcePlan
		r.CandidatePlan, r.CandidatePlanSHA256 = &plan, owner.scalar.prepared.PlanSHA256()
	}
	for _, prepared := range owner.criteria {
		entry := criterionSourceFlow(ctx, filename, source, prepared)
		if entry.ValueFlow.Status != "RESOLVED" {
			r.Status = "UNRESOLVED"
		}
		r.Predicates = append(r.Predicates, entry)
	}
	return r
}

func criterionSourceFlow(ctx context.Context, filename string, source []byte, prepared preparedRecordCriterion) CriterionSourceFlow {
	entry := CriterionSourceFlow{ID: prepared.declaration.ID, Activity: prepared.declaration.Activity,
		ActivityID: prepared.activityID, GeneratedSHA256: prepared.goSHA256}
	body, err := prepareActivityBodySource(filename, source, prepared.declaration.Activity)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		entry.ValueFlow = &RecordValueFlow{Schema: "gooo/fixed-predicate-value-flow/v1", Status: "UNRESOLVED", Reason: err.Error()}
		return entry
	}
	input := uint16(1)
	for i, parameter := range body.parameters {
		role := "original_input"
		if i == len(body.parameters)-1 {
			role = "candidate_result"
		}
		entry.Inputs = append(entry.Inputs, CriterionSourceInput{Input: input, Role: role, Type: parameter.Type})
		input++
	}
	plan := recordAssemblyPlan{body: body, spec: &assemblyspec.Spec{}}
	entry.ValueFlow = recordValueFlow(plan)
	entry.ValueFlow.Schema = "gooo/fixed-predicate-value-flow/v1"
	if entry.ValueFlow.Status == "RESOLVED" {
		graph, err := recordGraphInput(plan, entry.ValueFlow)
		if err != nil {
			entry.ValueFlow.Status, entry.ValueFlow.Reason = "UNRESOLVED", err.Error()
			entry.ValueFlow.Nodes, entry.ValueFlow.Choices, entry.ValueFlow.Helpers = nil, nil, nil
			return entry
		}
		entry.Nodes = graph.Nodes
		raw, _ := json.Marshal(entry.Nodes)
		entry.GraphSHA256 = digest(raw)
	}
	return entry
}
