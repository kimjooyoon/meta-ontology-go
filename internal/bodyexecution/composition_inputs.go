package bodyexecution

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func (node CompositionActivity) inputSlots() []CompositionInput {
	if len(node.Inputs) > 0 {
		return node.Inputs
	}
	return []CompositionInput{{Port: "input", Type: node.InputType, EntityID: node.InputEntityID, From: node.InputFrom}}
}

func (node CompositionActivity) inputKey(port string) string {
	if len(node.Inputs) == 0 {
		return node.Name
	}
	return node.Name + "." + port
}

func (graph *compositionGraph) bindActivity(file *syntax.File, i int) error {
	node := &graph.nodes[i]
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != node.Name {
			continue
		}
		if len(activity.Inputs) < 1 || len(activity.Inputs) > compositionLimit || !activity.ValueProgramPresent {
			return fmt.Errorf("composition activity %q requires 1..16 inputs and a computes body", node.Name)
		}
		node.InputType, node.OutputType = activity.Inputs[0].Name, activity.Output
		node.Assembling = activity.Assembly != nil
		if len(activity.Inputs) > 1 {
			node.Inputs = make([]CompositionInput, len(activity.Inputs))
			for p, input := range activity.Inputs {
				if graph.valueGoType(input.Name) == "" {
					return fmt.Errorf("composition activity %q requires a scalar or declared record", node.Name)
				}
				node.Inputs[p] = CompositionInput{Port: fmt.Sprintf("input%d", p), Type: input.Name, From: -1}
			}
		}
		if graph.valueGoType(node.InputType) == "" || graph.valueGoType(node.OutputType) == "" {
			return fmt.Errorf("composition activity %q requires a scalar or declared record", node.Name)
		}
		if node.Assembling && !bodycodegen.IsRecordAssembly(&activity.Assembly.Spec) &&
			(len(activity.Inputs) != 1 || node.InputType != "Integer" || node.OutputType != "Integer") {
			return fmt.Errorf("composition assembly activity %q requires Integer -> Integer", node.Name)
		}
		return nil
	}
	return fmt.Errorf("composition activity %q has no source declaration", node.ID)
}

func (graph *compositionGraph) bindEdges(typed bidir.TypedPlan) error {
	for _, edge := range typed.Edges {
		producer, consumer := graph.index(string(edge.SourceActivity)), graph.index(string(edge.TargetActivity))
		if producer < 0 || consumer < 0 || producer >= consumer || edge.SourcePort != "result" {
			return fmt.Errorf("composition edge must bind an earlier result to a declared input")
		}
		node := &graph.nodes[consumer]
		inputs := node.inputSlots()
		port, ok := semantic.InputPortIndex(edge.TargetPort, len(inputs))
		if !ok {
			return fmt.Errorf("composition input port %q is not declared", edge.TargetPort)
		}
		if inputs[port].From != -1 {
			return fmt.Errorf("composition activity %q port %q has multiple producers", node.Name, edge.TargetPort)
		}
		if len(node.Inputs) == 0 {
			node.InputFrom = producer
		} else {
			node.Inputs[port].From = producer
			if port == 0 {
				node.InputFrom = producer
			}
		}
		graph.plan.Edges = append(graph.plan.Edges, CompositionEdge{
			Producer: graph.nodes[producer].ID, Consumer: graph.nodes[consumer].ID,
			ProducerPort: "result", ConsumerPort: edge.TargetPort, Entity: inputs[port].EntityID})
	}
	return nil
}
