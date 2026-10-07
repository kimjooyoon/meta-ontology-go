package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func runtimeCompositionModel(file *syntax.File, ir semantic.IR, model generator.SemanticIR) (generator.SemanticIR, error) {
	if len(ir.RuntimeBindings) == 0 {
		return model, nil
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV4Support())
	if err != nil {
		return generator.SemanticIR{}, fmt.Errorf("document adaptation: %w", err)
	}
	plan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		return generator.SemanticIR{}, fmt.Errorf("typed plan: %w", err)
	}
	return appendRuntimeComposition(model, plan)
}

// appendRuntimeComposition projects only explicit, already type-checked bind
// edges into deterministic Go orchestration functions. It does not infer edges
// or claim that activity implementation slots perform domain work.
func appendRuntimeComposition(model generator.SemanticIR, plan bidir.TypedPlan) (generator.SemanticIR, error) {
	if len(plan.Edges) == 0 {
		return model, nil
	}
	activities := make(map[string]generator.Activity, len(model.Activities))
	for _, activity := range model.Activities {
		activities[activity.ID] = activity
	}
	adjacency := make(map[string][]string)
	for _, edge := range plan.Edges {
		adjacency[string(edge.SourceActivity)] = append(adjacency[string(edge.SourceActivity)], string(edge.TargetActivity))
		adjacency[string(edge.TargetActivity)] = append(adjacency[string(edge.TargetActivity)], string(edge.SourceActivity))
	}
	seen := make(map[string]bool)
	for _, activityID := range plan.Activities {
		root := string(activityID)
		if seen[root] || len(adjacency[root]) == 0 {
			continue
		}
		componentSet := map[string]bool{}
		queue := []string{root}
		seen[root] = true
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			componentSet[current] = true
			for _, next := range adjacency[current] {
				if !seen[next] {
					seen[next] = true
					queue = append(queue, next)
				}
			}
		}
		ordered := make([]string, 0, len(componentSet))
		for _, candidate := range plan.Activities {
			if componentSet[string(candidate)] {
				ordered = append(ordered, string(candidate))
			}
		}
		composition, err := runtimeCompositionActivity(ordered, activities, plan.Edges)
		if err != nil {
			return generator.SemanticIR{}, err
		}
		model.Activities = append(model.Activities, composition)
	}
	return model, nil
}

func runtimeCompositionActivity(ordered []string, activities map[string]generator.Activity, edges []bidir.BindingEdge) (generator.Activity, error) {
	component := make(map[string]bool, len(ordered))
	for _, id := range ordered {
		if _, ok := activities[id]; !ok {
			return generator.Activity{}, fmt.Errorf("typed plan activity %q is absent from Go projection", id)
		}
		component[id] = true
	}
	incoming := make(map[string]map[int]bidir.BindingEdge)
	outgoing := make(map[string]bool)
	sourceIndexes := make(map[string]int)
	var canonical strings.Builder
	for _, edge := range edges {
		if !component[string(edge.SourceActivity)] || !component[string(edge.TargetActivity)] {
			continue
		}
		sourceIndex, ok := runtimeBindingPortIndex(activities[string(edge.SourceActivity)].Outputs, edge.SourcePort, false)
		if !ok {
			return generator.Activity{}, fmt.Errorf("generated binding source port %q has no Go output on %q", edge.SourcePort, edge.SourceActivity)
		}
		targetIndex, ok := runtimeBindingPortIndex(activities[string(edge.TargetActivity)].Inputs, edge.TargetPort, true)
		if !ok {
			return generator.Activity{}, fmt.Errorf("generated binding target port %q has no Go input on %q", edge.TargetPort, edge.TargetActivity)
		}
		target := string(edge.TargetActivity)
		if incoming[target] == nil {
			incoming[target] = make(map[int]bidir.BindingEdge)
		}
		incoming[target][targetIndex] = edge
		sourceIndexes[runtimeCompositionEdgeKey(edge)] = sourceIndex
		outgoing[string(edge.SourceActivity)] = true
		fmt.Fprintf(&canonical, "%s:%s>%s:%s\n", edge.SourceActivity, edge.SourcePort, edge.TargetActivity, edge.TargetPort)
	}
	var name strings.Builder
	name.WriteString("GoooCompose")
	for _, id := range ordered {
		activity := activities[id]
		name.WriteString(activity.GoName)
		fmt.Fprintf(&canonical, "activity:%s\n", id)
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	id := "gooo://runtime-composition/v1/" + hex.EncodeToString(digest[:])
	var body strings.Builder
	inputs := make([]generator.Port, 0)
	for _, id := range ordered {
		activity := activities[id]
		for portIndex, port := range activity.Inputs {
			if _, bound := incoming[id][portIndex]; bound {
				continue
			}
			port.GoName = fmt.Sprintf("input%d", len(inputs))
			inputs = append(inputs, port)
		}
	}
	values := make(map[string][]string, len(ordered))
	inputIndex := 0
	for activityIndex, id := range ordered {
		activity := activities[id]
		args := make([]string, len(activity.Inputs))
		for portIndex := range activity.Inputs {
			if edge, bound := incoming[id][portIndex]; bound {
				sourceIndex := sourceIndexes[runtimeCompositionEdgeKey(edge)]
				args[portIndex] = values[string(edge.SourceActivity)][sourceIndex]
			} else {
				args[portIndex] = inputs[inputIndex].GoName
				inputIndex++
			}
		}
		if len(activity.Outputs) == 0 {
			fmt.Fprintf(&body, "%s(%s)\n", activity.GoName, strings.Join(args, ", "))
			continue
		}
		locals := make([]string, len(activity.Outputs))
		for outputIndex := range locals {
			locals[outputIndex] = fmt.Sprintf("runtime%d_%d", activityIndex, outputIndex)
		}
		fmt.Fprintf(&body, "%s := %s(%s)\n", strings.Join(locals, ", "), activity.GoName, strings.Join(args, ", "))
		values[id] = locals
	}
	outputs := make([]generator.Port, 0)
	for _, id := range ordered {
		if outgoing[id] {
			continue
		}
		for portIndex, output := range activities[id].Outputs {
			output.GoName = fmt.Sprintf("result%d", len(outputs))
			outputs = append(outputs, output)
			fmt.Fprintf(&body, "_ = %s\n", values[id][portIndex])
		}
	}
	if len(outputs) > 0 {
		// Return terminal activity values in the same deterministic traversal
		// used to construct output ports.
		terminal := make([]string, 0, len(outputs))
		for _, id := range ordered {
			if !outgoing[id] {
				terminal = append(terminal, values[id]...)
			}
		}
		fmt.Fprintf(&body, "return %s\n", strings.Join(terminal, ", "))
	}
	span := generator.SourceSpan{}
	for _, edge := range edges {
		if component[string(edge.SourceActivity)] && component[string(edge.TargetActivity)] {
			span = bidirGeneratorSpan(edge.Span)
			break
		}
	}
	return generator.Activity{
		ID: id, Name: name.String(), GoName: name.String(), Inputs: inputs, Outputs: outputs,
		Slots: []generator.Slot{{ID: id + "/implementation", Name: "implementation", Default: body.String(), Source: span}}, Source: span,
	}, nil
}

func runtimeCompositionEdgeKey(edge bidir.BindingEdge) string {
	return fmt.Sprintf("%s:%s>%s:%s", edge.SourceActivity, edge.SourcePort, edge.TargetActivity, edge.TargetPort)
}

func runtimeBindingPortIndex(ports []generator.Port, name string, input bool) (int, bool) {
	if input {
		if index, ok := semantic.InputPortIndex(name, len(ports)); ok {
			return index, true
		}
	} else if len(ports) == 1 && name == "result" {
		return 0, true
	}
	match := -1
	for index, port := range ports {
		if strings.EqualFold(port.Name, name) || strings.EqualFold(port.GoName, name) {
			if match >= 0 {
				return 0, false
			}
			match = index
		}
	}
	return match, match >= 0
}
