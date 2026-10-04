package bodyexecution

import (
	"context"
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const compositionLimit = 16

type CompositionActivity struct {
	Name           string             `json:"name"`
	ID             string             `json:"id"`
	InputType      string             `json:"input_type"`
	OutputType     string             `json:"output_type"`
	InputEntityID  string             `json:"input_entity_id"`
	OutputEntityID string             `json:"output_entity_id"`
	GoFunction     string             `json:"go_function"`
	InputFrom      int                `json:"input_from"`
	Assembling     bool               `json:"assembling"`
	Inputs         []CompositionInput `json:"inputs,omitempty"`
}

type CompositionInput struct {
	Port     string `json:"port"`
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
	From     int    `json:"from"`
}

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

type CompositionEdge struct {
	Producer     string `json:"producer"`
	Consumer     string `json:"consumer"`
	ProducerPort string `json:"producer_port"`
	ConsumerPort string `json:"consumer_port"`
	Entity       string `json:"entity"`
}

type CompositionPlan struct {
	Schema              string                `json:"schema"`
	TypedPlanSHA256     string                `json:"typed_plan_sha256"`
	SemanticFingerprint string                `json:"semantic_fingerprint"`
	Activities          []CompositionActivity `json:"activities"`
	Edges               []CompositionEdge     `json:"edges"`
}

type compositionGraph struct {
	plan  CompositionPlan
	nodes [compositionLimit]CompositionActivity
	count int
}

func prepareCompositionGraph(ctx context.Context, filename string, source []byte) (compositionGraph, error) {
	var graph compositionGraph
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return graph, fmt.Errorf("composition requires context and 1..128 KiB UTF-8 source")
	}
	if err := ctx.Err(); err != nil {
		return graph, err
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return graph, fmt.Errorf("composition source: %v", diagnostics)
	}
	for _, binding := range file.Bindings {
		if binding.Feedback {
			return graph, fmt.Errorf("composition requires in-invocation binds; feedback has a separate lifecycle")
		}
	}
	document, err := bidir.DocumentFromSyntaxContext(ctx, file)
	if err != nil {
		return graph, err
	}
	model, err := bidir.Get(document)
	if err != nil {
		return graph, err
	}
	typed, err := bidir.CompileTypedPlan(document)
	if err != nil {
		return graph, err
	}
	if len(typed.Activities) < 2 || len(typed.Activities) > compositionLimit {
		return graph, fmt.Errorf("composition requires 2..16 activities and explicit binds")
	}
	graph.count = len(typed.Activities)
	graph.plan = CompositionPlan{Schema: "gooo/body-composition-plan/v1",
		TypedPlanSHA256: typed.Digest(), SemanticFingerprint: bidir.SemanticFingerprint(model),
		Edges: make([]CompositionEdge, 0, len(typed.Edges))}
	for i, id := range typed.Activities {
		for _, node := range model.Nodes {
			if node.Kind == bidir.ActivityKind && node.ID == id {
				graph.nodes[i] = CompositionActivity{Name: node.Name, ID: string(id), InputFrom: -1,
					GoFunction: fmt.Sprintf("GoooComposedActivity%d", i)}
			}
		}
		if err := graph.bindActivity(file, i); err != nil {
			return graph, err
		}
		for _, entity := range model.Nodes {
			if entity.Kind != bidir.EntityKind {
				continue
			}
			if entity.Name == graph.nodes[i].InputType {
				graph.nodes[i].InputEntityID = string(entity.ID)
			}
			if entity.Name == graph.nodes[i].OutputType {
				graph.nodes[i].OutputEntityID = string(entity.ID)
			}
			for p := range graph.nodes[i].Inputs {
				if entity.Name == graph.nodes[i].Inputs[p].Type {
					graph.nodes[i].Inputs[p].EntityID = string(entity.ID)
				}
			}
		}
	}
	if err := graph.bindEdges(typed); err != nil {
		return graph, err
	}
	graph.plan.Activities = append([]CompositionActivity(nil), graph.nodes[:graph.count]...)
	return graph, nil
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
				if scalarGoType(input.Name) == "" {
					return fmt.Errorf("composition activity %q supports Integer, Boolean and Text", node.Name)
				}
				node.Inputs[p] = CompositionInput{Port: fmt.Sprintf("input%d", p), Type: input.Name, From: -1}
			}
		}
		if scalarGoType(node.InputType) == "" || scalarGoType(node.OutputType) == "" {
			return fmt.Errorf("composition activity %q supports Integer, Boolean and Text", node.Name)
		}
		if node.Assembling && (len(activity.Inputs) != 1 || node.InputType != "Integer" || node.OutputType != "Integer") {
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

func (graph *compositionGraph) index(id string) int {
	for i, node := range graph.nodes[:graph.count] {
		if node.ID == id {
			return i
		}
	}
	return -1
}

func scalarGoType(entity string) string {
	switch entity {
	case "Integer":
		return "int64"
	case "Boolean":
		return "bool"
	case "Text":
		return "string"
	}
	return ""
}

// Check all bodies and embedded plans before loading an optional model.
func (graph compositionGraph) preflight(ctx context.Context, filename string, source []byte) error {
	for _, node := range graph.nodes[:graph.count] {
		if _, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", ""); err != nil {
			return fmt.Errorf("activity %q: %w", node.Name, err)
		}
		if node.Assembling {
			if _, err := bodycodegen.DecodeSourcePathDocument(ctx, filename, source, node.Name, nil); err != nil {
				return fmt.Errorf("activity %q plan: %w", node.Name, err)
			}
		}
	}
	return nil
}

func sameCompositionPlan(left, right CompositionPlan) bool { return reflect.DeepEqual(left, right) }
