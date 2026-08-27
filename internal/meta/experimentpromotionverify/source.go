package experimentpromotionverify

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

var verifierEntities = map[string]bool{"ExperimentPortfolio": true, "PromotionGate": true, "ObservationReceipt": true}

// parseSource uses a verifier-specific vocabulary table and a map-backed
// declaration model. It repeats the parser/lowerer boundary but not the
// producer's graph traversal or identity implementation.
func parseSource(raw []byte) (SourceProjection, error) {
	projection, err := parseMaterial(raw)
	if err != nil {
		return projection, err
	}
	if len(projection.Experiments) != ExperimentCount || !sameIdentities(projection.Experiments, expectedIdentities()) || !sameStrings(projection.Gates, GateIDs) {
		return projection, fmt.Errorf("source vocabulary is not exact")
	}
	projection.Exact = true
	return projection, nil
}

func parseMaterial(raw []byte) (SourceProjection, error) {
	projection := SourceProjection{Path: SourcePath, RawDigest: digestBytes(raw)}
	file, diagnostics := syntax.ParseFile(SourcePath, string(raw))
	if file == nil || diagnostics.HasErrors() {
		return projection, fmt.Errorf("source syntax invalid")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projection, fmt.Errorf("source lowering invalid: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return projection, fmt.Errorf("source semantic IR invalid: %w", err)
	}
	entities := map[string]bool{}
	values := map[string]string{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Entity {
			entities[node.Name] = true
		}
		if node.Kind == semantic.Activity {
			if _, exists := values[node.Name]; exists {
				return projection, fmt.Errorf("duplicate activity")
			}
			values[node.Name] = node.ValueProgram
		}
	}
	if len(entities) != len(verifierEntities) || len(values) != 2 {
		return projection, fmt.Errorf("source declaration shape is not exact")
	}
	for entity := range verifierEntities {
		if !entities[entity] {
			return projection, fmt.Errorf("source entity is missing")
		}
	}
	portfolio, ok := values["DeclareExperimentPortfolio"]
	if !ok {
		return projection, fmt.Errorf("portfolio activity is missing")
	}
	gates, ok := values["DeclarePromotionGates"]
	if !ok {
		return projection, fmt.Errorf("gate activity is missing")
	}
	projection.Experiments, err = parseIdentityProgram(portfolio)
	if err != nil {
		return projection, err
	}
	projection.Gates, err = parseGateProgram(gates)
	if err != nil {
		return projection, err
	}
	if len(projection.Experiments) != ExperimentCount || len(projection.Gates) != GateCount {
		return projection, fmt.Errorf("source declaration cardinality is invalid")
	}
	projection.SemanticDigest = semanticDigest(projection.Experiments, projection.Gates)
	return projection, nil
}

func parseIdentityProgram(program string) ([]ExperimentIdentity, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != "experiments" {
		return nil, fmt.Errorf("experiment program head is invalid")
	}
	result := make([]ExperimentIdentity, 0, len(parts)-1)
	seenID := map[string]bool{}
	seenPR := map[int]bool{}
	for _, item := range parts[1:] {
		fields := strings.Split(item, "^")
		values := map[string]string{}
		for _, field := range fields {
			key, value, ok := strings.Cut(field, "=")
			if !ok || key == "" || value == "" || values[key] != "" {
				return nil, fmt.Errorf("experiment field is invalid")
			}
			values[key] = value
		}
		if len(values) != 4 {
			return nil, fmt.Errorf("experiment field count is invalid")
		}
		pr, err := strconv.Atoi(values["pr"])
		if err != nil || pr <= 0 || seenID[values["id"]] || seenPR[pr] {
			return nil, fmt.Errorf("IDENTITY_DUPLICATE")
		}
		seenID[values["id"]], seenPR[pr] = true, true
		result = append(result, ExperimentIdentity{ID: values["id"], PRNumber: pr, Topic: values["topic"], ClaimAddress: values["claim"]})
	}
	return result, nil
}

func parseGateProgram(program string) ([]string, error) {
	parts := strings.Split(program, "|")
	if len(parts) < 2 || parts[0] != "gates" {
		return nil, fmt.Errorf("gate program head is invalid")
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(parts)-1)
	for _, item := range parts[1:] {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key != "id" || value == "" || seen[value] {
			return nil, fmt.Errorf("gate identity is invalid")
		}
		seen[value] = true
		result = append(result, value)
	}
	return result, nil
}

func expectedIdentities() []ExperimentIdentity {
	topics := []string{"claim-lifecycle", "resolution-descent", "provenance", "capability", "phase-separation", "hygiene", "proof-choice", "counterexample-first", "causal-ci", "ambiguity", "freshness", "reproducibility", "observer", "partial-knowledge", "proof-artifact", "meta-circular", "reflective-sandbox", "policy", "invariant-transform", "semantic-delta", "audience", "claim-dependency", "refutation", "denominator", "external-oracle", "fixed-point", "resource", "quorum", "causal-explanation", "portfolio"}
	prs := []int{549, 548, 545, 555, 544, 543, 546, 559, 550, 552, 564, 547, 558, 542, 569, 567, 551, 560, 553, 562, 563, 566, 554, 570, 541, 556, 561, 568, 565, 557}
	result := make([]ExperimentIdentity, ExperimentCount)
	for i := range result {
		result[i] = ExperimentIdentity{ID: fmt.Sprintf("experiment-%02d", i+1), PRNumber: prs[i], Topic: topics[i], ClaimAddress: fmt.Sprintf("github://kimjooyoon/meta-ontology-go/pull/%d#%s", prs[i], topics[i])}
	}
	return result
}

func semanticDigest(experiments []ExperimentIdentity, gates []string) string {
	items := make([]string, 0, len(experiments))
	for _, value := range experiments {
		items = append(items, value.ID+"#"+strconv.Itoa(value.PRNumber)+"#"+value.Topic+"#"+value.ClaimAddress)
	}
	return digestBytes([]byte("experiment-promotion-semantic/v2|experiments=" + strings.Join(items, ",") + "|gates=" + strings.Join(gates, ",")))
}

func sameIdentities(left, right []ExperimentIdentity) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
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
