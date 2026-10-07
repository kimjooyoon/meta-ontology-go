package bodyexecution

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
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
	Prepared       bool               `json:"prepared,omitempty"`
	Inputs         []CompositionInput `json:"inputs,omitempty"`
}

type CompositionInput struct {
	Port     string `json:"port"`
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
	From     int    `json:"from"`
}

type CompositionEdge struct {
	Producer     string `json:"producer"`
	Consumer     string `json:"consumer"`
	ProducerPort string `json:"producer_port"`
	ConsumerPort string `json:"consumer_port"`
	Entity       string `json:"entity"`
}

type CompositionPlan struct {
	Schema              string                   `json:"schema"`
	EntryActivity       string                   `json:"entry_activity,omitempty"`
	TypedPlanSHA256     string                   `json:"typed_plan_sha256"`
	SemanticFingerprint string                   `json:"semantic_fingerprint"`
	Activities          []CompositionActivity    `json:"activities"`
	Edges               []CompositionEdge        `json:"edges"`
	Records             []bodycodegen.RecordType `json:"record_types,omitempty"`
	Preparations        []CompositionHelper      `json:"preparations,omitempty"`
}

type CompositionHelper struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type compositionGraph struct {
	plan     CompositionPlan
	nodes    [compositionLimit]CompositionActivity
	count    int
	deferred map[string]bool
	called   map[string]CompositionActivity
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

func sameCompositionPlan(left, right CompositionPlan) bool { return reflect.DeepEqual(left, right) }
