package experimentpromotion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{
	"ExperimentPortfolio", "PromotionGate", "ObservationReceipt",
}

var requiredActivities = []string{
	"DeclareExperimentPortfolio", "DeclarePromotionGates",
}

// parseSource is the producer's source authority. It consumes the raw .gooo,
// lowers it through the canonical parser/lowerer, and then decodes only the
// declared portfolio vocabulary. No outcome or claim label is accepted here.
func parseSource(raw []byte) (SourceProjection, error) {
	file, diagnostics := syntax.ParseFile(SourcePath, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return SourceProjection{Path: SourcePath, RawDigest: DigestBytes(raw)}, fmt.Errorf("GOOO_SOURCE_SYNTAX_INVALID")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceProjection{Path: SourcePath, RawDigest: DigestBytes(raw)}, fmt.Errorf("GOOO_SOURCE_LOWER_INVALID: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return SourceProjection{Path: SourcePath, RawDigest: DigestBytes(raw)}, fmt.Errorf("GOOO_SOURCE_IR_INVALID: %w", err)
	}

	entities := make([]string, 0)
	activities := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			entities = append(entities, node.Name)
		case semantic.Activity:
			if _, exists := activities[node.Name]; exists {
				return SourceProjection{}, fmt.Errorf("GOOO_SOURCE_DUPLICATE_ACTIVITY")
			}
			activities[node.Name] = node.ValueProgram
		}
	}
	sort.Strings(entities)
	projection := SourceProjection{
		Path:           SourcePath,
		RawDigest:      DigestBytes(raw),
		SemanticDigest: "",
		Experiments:    nil,
		Gates:          nil,
	}
	projection.Exact = exactNames(entities, requiredEntities) && len(activities) == len(requiredActivities)
	for _, name := range requiredActivities {
		if _, ok := activities[name]; !ok {
			return projection, fmt.Errorf("GOOO_SOURCE_ACTIVITY_MISSING: %s", name)
		}
	}
	var errExperiments, errGates error
	projection.Experiments, errExperiments = decodeIDs(activities["DeclareExperimentPortfolio"], "experiments")
	projection.Gates, errGates = decodeIDs(activities["DeclarePromotionGates"], "gates")
	if errExperiments != nil {
		return projection, errExperiments
	}
	if errGates != nil {
		return projection, errGates
	}
	projection.Exact = projection.Exact && sameStrings(projection.Experiments, experimentIDs()) && sameStrings(projection.Gates, GateIDs)
	if !projection.Exact {
		return projection, fmt.Errorf("GOOO_SOURCE_VOCABULARY_NOT_EXACT")
	}
	projection.SemanticDigest = semanticDigest(projection.Experiments, projection.Gates)
	return projection, nil
}

func semanticDigest(experiments, gates []string) string {
	canonical := "experiment-promotion-semantic/v1|entities=ExperimentPortfolio,ObservationReceipt,PromotionGate|activity=DeclareExperimentPortfolio:" + strings.Join(experiments, ",") + "|activity=DeclarePromotionGates:" + strings.Join(gates, ",")
	return DigestBytes([]byte(canonical))
}

func decodeIDs(value, expectedHead string) ([]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || parts[0] != expectedHead {
		return nil, fmt.Errorf("GOOO_SOURCE_PAYLOAD_INVALID: %s", expectedHead)
	}
	result := make([]string, 0, len(parts)-1)
	seen := make(map[string]bool)
	for _, part := range parts[1:] {
		key, id, ok := strings.Cut(part, "=")
		if !ok || key != "id" || id == "" || seen[id] {
			return nil, fmt.Errorf("GOOO_SOURCE_ID_FIELD_INVALID: %s", part)
		}
		seen[id] = true
		result = append(result, id)
	}
	return result, nil
}

func experimentIDs() []string {
	result := make([]string, ExperimentCount)
	for i := range result {
		result[i] = fmt.Sprintf("experiment-%02d", i+1)
	}
	return result
}

func exactNames(observed, expected []string) bool {
	if len(observed) != len(expected) {
		return false
	}
	return sameStrings(observed, sortedCopy(expected))
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
