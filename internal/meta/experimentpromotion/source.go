package experimentpromotion

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var requiredEntities = []string{"ExperimentPortfolio", "PromotionGate", "ObservationReceipt"}

// parseSource is the producer's source authority. It consumes raw .gooo,
// lowers it through the canonical parser/lowerer, and decodes only identity
// material. It never reads a state, decision, or PASS label from source.
func parseSource(raw []byte) (SourceProjection, error) {
	projection, err := parseSourceMaterial(raw)
	if err != nil {
		return projection, err
	}
	if !projection.Exact {
		return projection, fmt.Errorf("GOOO_SOURCE_VOCABULARY_NOT_EXACT")
	}
	if err := validateIdentityShape(projection.Experiments); err != nil {
		return projection, err
	}
	return projection, nil
}

// parseSourceMaterial accepts a well-formed portfolio-shaped intervention as
// well as the canonical portfolio. This lets semantic-causality compare real
// source mutations rather than compare random digest metadata.
func parseSourceMaterial(raw []byte) (SourceProjection, error) {
	file, diagnostics := syntax.ParseFile(SourcePath, string(raw))
	projection := SourceProjection{Path: SourcePath, RawDigest: DigestBytes(raw)}
	if file == nil || diagnostics.HasErrors() {
		return projection, fmt.Errorf("GOOO_SOURCE_SYNTAX_INVALID")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projection, fmt.Errorf("GOOO_SOURCE_LOWER_INVALID: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return projection, fmt.Errorf("GOOO_SOURCE_IR_INVALID: %w", err)
	}

	entities := make([]string, 0)
	activities := make(map[string]string)
	for _, node := range ir.Graph.Nodes() {
		switch node.Kind {
		case semantic.Entity:
			entities = append(entities, node.Name)
		case semantic.Activity:
			if _, exists := activities[node.Name]; exists {
				return projection, fmt.Errorf("GOOO_SOURCE_DUPLICATE_ACTIVITY")
			}
			activities[node.Name] = node.ValueProgram
		}
	}
	sort.Strings(entities)
	if !sameStrings(entities, sortedCopy(requiredEntities)) || len(activities) != 2 {
		return projection, fmt.Errorf("GOOO_SOURCE_DECLARATION_SHAPE_INVALID")
	}
	portfolio, ok := activities["DeclareExperimentPortfolio"]
	if !ok {
		return projection, fmt.Errorf("GOOO_SOURCE_ACTIVITY_MISSING: DeclareExperimentPortfolio")
	}
	gates, ok := activities["DeclarePromotionGates"]
	if !ok {
		return projection, fmt.Errorf("GOOO_SOURCE_ACTIVITY_MISSING: DeclarePromotionGates")
	}
	projection.Experiments, err = decodeExperiments(portfolio)
	if err != nil {
		return projection, err
	}
	projection.Gates, err = decodeIDs(gates, "gates")
	if err != nil {
		return projection, err
	}
	if len(projection.Experiments) != ExperimentCount || !sameStrings(projection.Gates, GateIDs) {
		return projection, fmt.Errorf("GOOO_SOURCE_VOCABULARY_NOT_EXACT")
	}
	projection.SemanticDigest = semanticDigest(projection.Experiments, projection.Gates)
	projection.Exact = true
	return projection, nil
}

func decodeExperiments(value string) ([]ExperimentIdentity, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || parts[0] != "experiments" {
		return nil, fmt.Errorf("GOOO_SOURCE_EXPERIMENT_PAYLOAD_INVALID")
	}
	result := make([]ExperimentIdentity, 0, len(parts)-1)
	seenID := make(map[string]bool)
	seenPR := make(map[int]bool)
	for _, part := range parts[1:] {
		fields := strings.Split(part, "^")
		if len(fields) != 4 {
			return nil, fmt.Errorf("GOOO_SOURCE_EXPERIMENT_ARITY_INVALID")
		}
		values := make(map[string]string)
		for _, field := range fields {
			key, value, ok := strings.Cut(field, "=")
			if !ok || value == "" || values[key] != "" {
				return nil, fmt.Errorf("GOOO_SOURCE_EXPERIMENT_FIELD_INVALID")
			}
			values[key] = value
		}
		if len(values) != 4 || values["id"] == "" || values["pr"] == "" || values["topic"] == "" || values["claim"] == "" {
			return nil, fmt.Errorf("GOOO_SOURCE_EXPERIMENT_FIELDS_INCOMPLETE")
		}
		pr, err := strconv.Atoi(values["pr"])
		if err != nil || pr <= 0 || seenID[values["id"]] || seenPR[pr] {
			return nil, fmt.Errorf("GOOO_SOURCE_EXPERIMENT_IDENTITY_DUPLICATE")
		}
		seenID[values["id"]], seenPR[pr] = true, true
		result = append(result, ExperimentIdentity{ID: values["id"], PRNumber: pr, Topic: values["topic"], ClaimAddress: values["claim"]})
	}
	return result, nil
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

func validateIdentityShape(values []ExperimentIdentity) error {
	if len(values) != ExperimentCount {
		return fmt.Errorf("GOOO_SOURCE_EXPERIMENT_COUNT_INVALID")
	}
	seenPR := make(map[int]bool)
	for index, value := range values {
		if value.ID != fmt.Sprintf("experiment-%02d", index+1) || value.PRNumber <= 0 || value.Topic == "" || value.ClaimAddress == "" || seenPR[value.PRNumber] {
			return fmt.Errorf("GOOO_SOURCE_EXPERIMENT_IDENTITY_INVALID: %s", value.ID)
		}
		seenPR[value.PRNumber] = true
	}
	return nil
}

func semanticDigest(experiments []ExperimentIdentity, gates []string) string {
	parts := make([]string, 0, len(experiments))
	for _, value := range experiments {
		parts = append(parts, fmt.Sprintf("%s#%d#%s#%s", value.ID, value.PRNumber, value.Topic, value.ClaimAddress))
	}
	canonical := "experiment-promotion-semantic/v2|experiments=" + strings.Join(parts, ",") + "|gates=" + strings.Join(gates, ",")
	return DigestBytes([]byte(canonical))
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
