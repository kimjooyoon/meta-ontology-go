package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

type intervention struct {
	Name                       string   `json:"name"`
	Kind                       string   `json:"kind"`
	BaselineSourceDigest       string   `json:"baseline_source_digest"`
	InterventionSourceDigest   string   `json:"intervention_source_digest"`
	BaselineSemanticDigest     string   `json:"baseline_semantic_digest"`
	InterventionSemanticDigest string   `json:"intervention_semantic_digest"`
	BaselineGraphDigest        string   `json:"baseline_graph_digest"`
	InterventionGraphDigest    string   `json:"intervention_graph_digest"`
	BaselineDecision           string   `json:"baseline_decision"`
	InterventionDecision       string   `json:"intervention_decision"`
	BaselineStates             []string `json:"baseline_states"`
	InterventionStates         []string `json:"intervention_states"`
	BaselineCausePath          []string `json:"baseline_cause_path"`
	InterventionCausePath      []string `json:"intervention_cause_path"`
	SemanticDigestChanged      bool     `json:"semantic_digest_changed"`
	GraphDigestChanged         bool     `json:"graph_digest_changed"`
	StateTransitionChanged     bool     `json:"state_transition_changed"`
	CausePathChanged           bool     `json:"cause_path_changed"`
	DecisionChanged            bool     `json:"decision_changed"`
	EdgeTypeChanged            bool     `json:"edge_type_changed"`
}

type report struct {
	Schema           string         `json:"schema"`
	ReadOnly         bool           `json:"read_only"`
	RepositoryWrites int            `json:"repository_writes"`
	Interventions    []intervention `json:"interventions"`
}

func main() {
	baselinePath := flag.String("baseline", "", "baseline Gooo source")
	semanticPath := flag.String("semantic", "", "semantic-value intervention source")
	edgePath := flag.String("edge", "", "edge-type intervention source")
	commentPath := flag.String("comment", "", "comment-only intervention source")
	outputPath := flag.String("output", "", "intervention artifact output path")
	flag.Parse()
	if *baselinePath == "" || *semanticPath == "" || *edgePath == "" || *commentPath == "" || *outputPath == "" {
		fail("-baseline, -semantic, -edge, -comment, and -output are required")
	}
	baseline := read(*baselinePath)
	semanticSource := read(*semanticPath)
	edgeSource := read(*edgePath)
	commentSource := read(*commentPath)
	items := []intervention{
		compare("semantic-value", "VALUE_PROGRAM", baseline, *baselinePath, semanticSource, *semanticPath, claimdependency.ObservationUnknown, claimdependency.ObservationContradiction),
		compare("edge-type", "EDGE_KIND", baseline, *baselinePath, edgeSource, *edgePath, claimdependency.ObservationUnknown, claimdependency.ObservationContradiction),
		compare("comment-only", "COMMENT_ONLY", baseline, *baselinePath, commentSource, *commentPath, claimdependency.ObservationUnknown, claimdependency.ObservationUnknown),
	}
	result := report{Schema: "gooo.meta.claim-dependency-intervention/v1", ReadOnly: true, RepositoryWrites: 0, Interventions: items}
	writeJSON(*outputPath, result)
	for _, item := range items {
		fmt.Printf("intervention=%s semantic_digest_changed=%t graph_digest_changed=%t state_transition_changed=%t cause_path_changed=%t decision_changed=%t edge_type_changed=%t\n", item.Name, item.SemanticDigestChanged, item.GraphDigestChanged, item.StateTransitionChanged, item.CausePathChanged, item.DecisionChanged, item.EdgeTypeChanged)
	}
}

func compare(name, kind string, baseline []byte, baselinePath string, changed []byte, changedPath string, baselinePredicate, changedPredicate claimdependency.ObservationPredicate) intervention {
	baseObservation, err := claimdependency.ObservationForSource(baseline, baselinePath, baselinePredicate, evidenceFor(baselinePredicate))
	if err != nil {
		fail(err.Error())
	}
	changedObservation, err := claimdependency.ObservationForSource(changed, changedPath, changedPredicate, evidenceFor(changedPredicate))
	if err != nil {
		fail(err.Error())
	}
	baseReceipt, err := claimdependency.Evaluate(baseline, baselinePath, baseObservation, nil)
	if err != nil {
		fail(err.Error())
	}
	changedReceipt, err := claimdependency.Evaluate(changed, changedPath, changedObservation, nil)
	if err != nil {
		fail(err.Error())
	}
	baseStates, changedStates := states(baseReceipt), states(changedReceipt)
	basePath, changedPathValue := baseReceipt.Resolutions[len(baseReceipt.Resolutions)-1].CausePath, changedReceipt.Resolutions[len(changedReceipt.Resolutions)-1].CausePath
	return intervention{
		Name: name, Kind: kind,
		BaselineSourceDigest: baseReceipt.Subject.SourceDigest, InterventionSourceDigest: changedReceipt.Subject.SourceDigest,
		BaselineSemanticDigest: baseReceipt.Subject.SemanticDigest, InterventionSemanticDigest: changedReceipt.Subject.SemanticDigest,
		BaselineGraphDigest: baseReceipt.Graph.Digest, InterventionGraphDigest: changedReceipt.Graph.Digest,
		BaselineDecision: baseReceipt.Decision.Value + ":" + baseReceipt.Decision.Resolution, InterventionDecision: changedReceipt.Decision.Value + ":" + changedReceipt.Decision.Resolution,
		BaselineStates: baseStates, InterventionStates: changedStates, BaselineCausePath: basePath, InterventionCausePath: changedPathValue,
		SemanticDigestChanged:  baseReceipt.Subject.SemanticDigest != changedReceipt.Subject.SemanticDigest,
		GraphDigestChanged:     baseReceipt.Graph.Digest != changedReceipt.Graph.Digest,
		StateTransitionChanged: !reflect.DeepEqual(baseStates, changedStates), CausePathChanged: !reflect.DeepEqual(basePath, changedPathValue),
		DecisionChanged: baseReceipt.Decision != changedReceipt.Decision,
		EdgeTypeChanged: !reflect.DeepEqual(baseReceipt.Graph.Edges, changedReceipt.Graph.Edges),
	}
}

func evidenceFor(predicate claimdependency.ObservationPredicate) string {
	if predicate == claimdependency.ObservationUnknown {
		return ""
	}
	return "controlled evidence predicate"
}
func states(receipt claimdependency.Receipt) []string {
	result := make([]string, len(receipt.Resolutions))
	for i, value := range receipt.Resolutions {
		result[i] = value.State
	}
	return result
}
func read(path string) []byte {
	value, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	return value
}
func writeJSON(path string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fail(err.Error())
	}
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(2) }
