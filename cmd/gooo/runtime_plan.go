package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const runtimePlanSchema = "gooo/runtime-plan/v1"

type runtimePlanDocument struct {
	Schema              string               `json:"schema"`
	SourceDigest        string               `json:"source_digest"`
	SemanticHash        string               `json:"semantic_hash"`
	TypedPlanDigest     string               `json:"typed_plan_digest,omitempty"`
	ActivityOrder       []string             `json:"activity_order,omitempty"`
	BindingEdgeOrder    []string             `json:"binding_edge_order,omitempty"`
	RuntimeBindingCount int                  `json:"runtime_binding_count"`
	Bindings            []runtimePlanBinding `json:"bindings"`
}

type runtimePlanBinding struct {
	Schema           string `json:"schema"`
	ProducerActivity string `json:"producer_activity"`
	ProducerPort     string `json:"producer_port"`
	ConsumerActivity string `json:"consumer_activity"`
	ConsumerPort     string `json:"consumer_port"`
	Entity           string `json:"entity"`
}

func runtimePlanBindingEdgeKey(edge bidir.BindingEdge) string {
	return fmt.Sprintf("%s:%s->%s:%s", edge.SourceActivity, edge.SourcePort, edge.TargetActivity, edge.TargetPort)
}

func buildRuntimePlanData(source []byte, ir semantic.IR) ([]byte, error) {
	return buildRuntimePlanDataWithTypedPlan(source, ir, bidir.TypedPlan{})
}

func runtimePlanSourceDigest(source []byte) string {
	return "sha256:" + cache.HashBytes(source).String()
}

func buildRuntimePlanDataWithTypedPlan(source []byte, ir semantic.IR, typedPlan bidir.TypedPlan) ([]byte, error) {
	bindings := make([]runtimePlanBinding, 0, len(ir.RuntimeBindings))
	for _, binding := range ir.RuntimeBindings {
		bindings = append(bindings, runtimePlanBinding{
			Schema: binding.Schema, ProducerActivity: string(binding.ProducerActivity), ProducerPort: binding.ProducerPort,
			ConsumerActivity: string(binding.ConsumerActivity), ConsumerPort: binding.ConsumerPort, Entity: string(binding.Entity),
		})
	}
	sort.Slice(bindings, func(i, j int) bool {
		left, right := bindings[i], bindings[j]
		if left.ProducerActivity != right.ProducerActivity {
			return left.ProducerActivity < right.ProducerActivity
		}
		if left.ConsumerActivity != right.ConsumerActivity {
			return left.ConsumerActivity < right.ConsumerActivity
		}
		if left.ProducerPort != right.ProducerPort {
			return left.ProducerPort < right.ProducerPort
		}
		return left.ConsumerPort < right.ConsumerPort
	})
	document := runtimePlanDocument{
		Schema: runtimePlanSchema, SourceDigest: runtimePlanSourceDigest(source), SemanticHash: ir.StableHash(),
		RuntimeBindingCount: len(bindings), Bindings: bindings,
	}
	if len(typedPlan.Activities) > 0 {
		document.TypedPlanDigest = typedPlan.Digest()
		document.ActivityOrder = make([]string, len(typedPlan.Activities))
		for index, activity := range typedPlan.Activities {
			document.ActivityOrder[index] = string(activity)
		}
		document.BindingEdgeOrder = make([]string, len(typedPlan.Edges))
		for index, edge := range typedPlan.Edges {
			document.BindingEdgeOrder[index] = runtimePlanBindingEdgeKey(edge)
		}
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
