package experimentpromotionverify

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{"ExperimentPortfolio", "PromotionGate", "ObservationReceipt"}

func parseSource(raw []byte) (SourceProjection, error) {
	file, diagnostics := syntax.ParseFile(SourcePath, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return SourceProjection{Path: SourcePath, RawDigest: digestBytes(raw)}, fmt.Errorf("source syntax invalid")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return SourceProjection{Path: SourcePath, RawDigest: digestBytes(raw)}, fmt.Errorf("source lowering invalid: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return SourceProjection{Path: SourcePath, RawDigest: digestBytes(raw)}, fmt.Errorf("source semantic IR invalid: %w", err)
	}
	entities := make([]string, 0)
	activities := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			entities = append(entities, node.Name)
		case semantic.Activity:
			if _, exists := activities[node.Name]; exists {
				return SourceProjection{}, fmt.Errorf("duplicate activity")
			}
			activities[node.Name] = node.ValueProgram
		}
	}
	sort.Strings(entities)
	projection := SourceProjection{Path: SourcePath, RawDigest: digestBytes(raw)}
	if len(entities) != len(requiredEntities) || !sameStrings(entities, sortedCopy(requiredEntities)) || len(activities) != 2 {
		return projection, fmt.Errorf("source declaration shape is not exact")
	}
	portfolio, ok := activities["DeclareExperimentPortfolio"]
	if !ok {
		return projection, fmt.Errorf("portfolio activity is missing")
	}
	gates, ok := activities["DeclarePromotionGates"]
	if !ok {
		return projection, fmt.Errorf("gate activity is missing")
	}
	projection.Experiments, err = decodeIDs(portfolio, "experiments")
	if err != nil {
		return projection, err
	}
	projection.Gates, err = decodeIDs(gates, "gates")
	if err != nil {
		return projection, err
	}
	projection.Exact = sameStrings(projection.Experiments, experimentIDs()) && sameStrings(projection.Gates, GateIDs)
	if !projection.Exact {
		return projection, fmt.Errorf("source vocabulary is not exact")
	}
	projection.SemanticDigest = semanticDigest(projection.Experiments, projection.Gates)
	return projection, nil
}

func semanticDigest(experiments, gates []string) string {
	canonical := "experiment-promotion-semantic/v1|entities=ExperimentPortfolio,ObservationReceipt,PromotionGate|activity=DeclareExperimentPortfolio:" + strings.Join(experiments, ",") + "|activity=DeclarePromotionGates:" + strings.Join(gates, ",")
	return digestBytes([]byte(canonical))
}

func decodeIDs(value, head string) ([]string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || parts[0] != head {
		return nil, fmt.Errorf("source payload head is invalid")
	}
	result := make([]string, 0, len(parts)-1)
	seen := make(map[string]bool)
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key != "id" || value == "" || seen[value] {
			return nil, fmt.Errorf("source id field is invalid")
		}
		seen[value] = true
		result = append(result, value)
	}
	return result, nil
}

func experimentIDs() []string {
	result := make([]string, ExperimentCount)
	for index := range result {
		result[index] = fmt.Sprintf("experiment-%02d", index+1)
	}
	return result
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

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
