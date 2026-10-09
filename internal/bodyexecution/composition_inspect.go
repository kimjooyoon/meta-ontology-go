package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// CompositionInspection exposes the same structural plan used by generation.
// Counters describe this inspection only; bodies and candidates are not tested.
type CompositionInspection struct {
	Schema           string                     `json:"schema"`
	SourceSHA256     string                     `json:"source_sha256"`
	Plan             CompositionPlan            `json:"plan"`
	CallerInputs     []CompositionCallerInput   `json:"caller_inputs"`
	Assemblies       []CompositionAssemblyInput `json:"assemblies"`
	ModelCalls       int                        `json:"model_calls"`
	CandidateTests   int                        `json:"candidate_tests"`
	NativeExecutions int                        `json:"native_executions"`
	Scope            string                     `json:"scope"`
}

type CompositionCallerInput struct {
	Key        string `json:"key"`
	ActivityID string `json:"activity_id"`
	Port       string `json:"port"`
	Type       string `json:"type"`
	ScalarKind string `json:"scalar_kind,omitempty"`
	EntityID   string `json:"entity_id"`
}

type CompositionAssemblyInput struct {
	Name               string `json:"name"`
	ID                 string `json:"id"`
	Phase              string `json:"phase"`
	Kind               string `json:"kind"`
	ContractSHA256     string `json:"contract_sha256"`
	SourceCases        int    `json:"source_cases"`
	SourceHoldoutCases int    `json:"source_holdout_cases"`
	DependsOnAssembly  bool   `json:"depends_on_assembly"`
}

// InspectComposition resolves roots, binds and called assembly prerequisites
// without loading models, realizing candidates, invoking Go or running bodies.
func InspectComposition(ctx context.Context, filename string, source []byte, entry string) (CompositionInspection, error) {
	r := CompositionInspection{Schema: "gooo/body-composition-inspection/v1", SourceSHA256: digest(source),
		CallerInputs: []CompositionCallerInput{}, Assemblies: []CompositionAssemblyInput{},
		Scope: "structural input plan; body type checks, candidate selection, model compatibility and correctness are separate"}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, entry)
	if err != nil {
		return r, err
	}
	r.Plan = graph.plan
	for _, node := range graph.nodes[:graph.count] {
		for _, input := range node.inputSlots() {
			if input.From < 0 {
				kind := graph.scalarKind(input.Type)
				if scalarGoType(kind) == "" {
					kind = ""
				}
				r.CallerInputs = append(r.CallerInputs, CompositionCallerInput{Key: node.inputKey(input.Port),
					ActivityID: node.ID, Port: input.Port, Type: input.Type, EntityID: input.EntityID, ScalarKind: kind})
			}
		}
	}
	for _, helper := range graph.plan.Preparations {
		if err := r.addAssembly(ctx, filename, source, helper.Name, helper.ID, "called_body", graph.deferred); err != nil {
			return r, err
		}
	}
	for _, node := range graph.nodes[:graph.count] {
		if node.Assembling && !node.Prepared {
			if err := r.addAssembly(ctx, filename, source, node.Name, node.ID, "activity", graph.deferred); err != nil {
				return r, err
			}
		}
	}
	return r, ctx.Err()
}

func (r *CompositionInspection) addAssembly(ctx context.Context, filename string, source []byte,
	name, id, phase string, deferred map[string]bool) error {
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, name)
	if err != nil {
		return err
	}
	if spec == nil {
		return fmt.Errorf("assembly %q has no source contract", name)
	}
	contract, err := spec.Canonical()
	if err != nil {
		return err
	}
	kind := "typed_paths"
	switch {
	case spec.FillPlan != nil:
		kind = "source_fill"
	case spec.Search != nil:
		kind = "source_search"
	case bodycodegen.IsRecordAssembly(spec):
		kind = "record_choices"
	}
	r.Assemblies = append(r.Assemblies, CompositionAssemblyInput{Name: name, ID: id, Phase: phase, Kind: kind,
		ContractSHA256: digest([]byte(contract)), SourceCases: len(spec.Cases) + len(spec.ValueCases),
		SourceHoldoutCases: len(spec.HoldoutCases) + len(spec.ValueHoldoutCases), DependsOnAssembly: deferred[name]})
	return nil
}

// CompositionInputTemplate emits one type-shaped input row, with no expectations.
// Zeros are editable placeholders; they are not inferred caller or source cases.
func CompositionInputTemplate(plan CompositionInspection) ([]byte, error) {
	if plan.Schema != "gooo/body-composition-inspection/v1" || len(plan.CallerInputs) < 1 ||
		len(plan.CallerInputs) > compositionLimit*compositionLimit {
		return nil, fmt.Errorf("input template requires a bounded composition inspection")
	}
	row := make(map[string]any, len(plan.CallerInputs))
	for _, input := range plan.CallerInputs {
		if _, exists := row[input.Key]; exists || input.Key == "" {
			return nil, fmt.Errorf("input template requires unique caller keys")
		}
		value, err := compositionInputPlaceholder(input, plan.Plan.Records)
		if err != nil {
			return nil, err
		}
		row[input.Key] = value
	}
	raw, err := json.MarshalIndent(struct {
		Schema string           `json:"schema"`
		Inputs []map[string]any `json:"inputs"`
	}{CompositionInputsSchema, []map[string]any{row}}, "", "  ")
	if err == nil && len(raw) > 32<<10 {
		err = fmt.Errorf("input template exceeds 32 KiB")
	}
	return raw, err
}

func compositionInputPlaceholder(input CompositionCallerInput, records []bodycodegen.RecordType) (any, error) {
	if value, ok := scalarInputPlaceholder(input.ScalarKind); ok {
		return value, nil
	}
	for _, record := range records {
		if record.ID != input.EntityID {
			continue
		}
		value := make(map[string]any, len(record.Fields))
		for _, field := range record.Fields {
			if field.Presence == "optional" {
				continue
			}
			item, ok := scalarInputPlaceholder(field.TypeID)
			if !ok {
				return nil, fmt.Errorf("input %q field %q has no template type", input.Key, field.Name)
			}
			value[field.Name] = item
		}
		return value, nil
	}
	return nil, fmt.Errorf("input %q has no template type", input.Key)
}

func scalarInputPlaceholder(kind string) (any, bool) {
	switch kind {
	case "Integer", "urn:gooo:type:integer":
		return int64(0), true
	case "Boolean", "urn:gooo:type:boolean":
		return false, true
	case "Text", "urn:gooo:type:string":
		return "", true
	}
	return nil, false
}
