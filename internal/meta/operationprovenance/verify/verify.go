package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const canonicalSource = `package operationprovenance
namespace operationprovenance

entity Metric id "gooo://meta-operation-provenance/entity/metric"
entity Producer id "gooo://meta-operation-provenance/entity/producer"
entity Consumer id "gooo://meta-operation-provenance/entity/consumer"
entity MetaOperation id "gooo://meta-operation-provenance/entity/meta-operation"
entity EvidencePath id "gooo://meta-operation-provenance/entity/evidence-path"
entity ProducerBoundMetric id "gooo://meta-operation-provenance/entity/producer-bound-metric"
entity ConsumerBoundMetric id "gooo://meta-operation-provenance/entity/consumer-bound-metric"
entity OperationBoundMetric id "gooo://meta-operation-provenance/entity/operation-bound-metric"
entity LineageBoundMetric id "gooo://meta-operation-provenance/entity/lineage-bound-metric"
entity FoundationMetricDecision id "gooo://meta-operation-provenance/entity/foundation-decision"
entity CoherenceMetricDecision id "gooo://meta-operation-provenance/entity/coherence-decision"
entity RegressionMetricDecision id "gooo://meta-operation-provenance/entity/regression-decision"
entity ClaimState id "gooo://meta-operation-provenance/entity/claim-state"
entity ProvenanceReceipt id "gooo://meta-operation-provenance/entity/receipt"

activity BindProducer(Metric) -> ProducerBoundMetric
activity BindConsumer(ProducerBoundMetric) -> ConsumerBoundMetric
activity BindMetaOperation(ConsumerBoundMetric) -> OperationBoundMetric
activity BindEvidencePath(OperationBoundMetric) -> LineageBoundMetric
activity DecideFoundation(LineageBoundMetric) -> FoundationMetricDecision
activity DecideCoherence(LineageBoundMetric) -> CoherenceMetricDecision
activity DecideRegression(LineageBoundMetric) -> RegressionMetricDecision
activity PreserveClaimState(FoundationMetricDecision) -> ClaimState
activity EmitProvenanceReceipt(ClaimState) -> ProvenanceReceipt
`

type receipt struct {
	Schema                    string     `json:"schema"`
	Toolchain                 string     `json:"toolchain"`
	RepositoryWorkspaceWrites bool       `json:"repository_workspace_writes"`
	MutationAuthority         bool       `json:"mutation_authority"`
	SourceDigest              string     `json:"source_digest"`
	Scenarios                 []scenario `json:"scenarios"`
	Digest                    string     `json:"digest"`
}
type scenario struct {
	ID          string         `json:"id"`
	Graph       graph          `json:"graph"`
	Numerator   int            `json:"numerator"`
	Denominator int            `json:"denominator"`
	Decisions   map[string]int `json:"decisions"`
	Metrics     []metric       `json:"metrics"`
}
type graph struct {
	Nodes     int            `json:"nodes"`
	Edges     int            `json:"edges"`
	EdgeKinds map[string]int `json:"edge_kinds"`
}
type metric struct {
	ID              string  `json:"id"`
	Family          string  `json:"family"`
	Claim           string  `json:"claim"`
	Numerator       int     `json:"numerator"`
	Denominator     int     `json:"denominator"`
	Decision        string  `json:"decision"`
	EvaluationState string  `json:"evaluation_state"`
	Lineage         lineage `json:"lineage"`
	Issue           *issue  `json:"issue,omitempty"`
}
type lineage struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	EvidencePath  string `json:"evidence_path"`
}
type issue struct {
	Stage     string   `json:"stage"`
	Step      string   `json:"step"`
	Reason    string   `json:"reason"`
	Cause     string   `json:"cause"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

func Verify(payload, source []byte) (map[string]any, error) {
	var actual receipt
	if err := decodeExact(payload, &actual); err != nil {
		return nil, fmt.Errorf("decode receipt: %w", err)
	}
	if !bytes.Equal(source, []byte(canonicalSource)) {
		return nil, fmt.Errorf("source differs from independent reconstruction")
	}
	file, diagnostics := syntax.ParseFile("main.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		return nil, fmt.Errorf("source does not parse independently")
	}
	if actual.Schema != "gooo/meta-operation-provenance-receipt/v1" || actual.Toolchain != "go1.27.0" || actual.RepositoryWorkspaceWrites || actual.MutationAuthority {
		return nil, fmt.Errorf("receipt mutation boundary is invalid")
	}
	if actual.SourceDigest != digest(source) || len(actual.Scenarios) != 4 {
		return nil, fmt.Errorf("receipt source binding or scenario count is invalid")
	}
	for _, scenario := range actual.Scenarios {
		if err := verifyScenario(scenario); err != nil {
			return nil, err
		}
	}
	withoutDigest := actual
	withoutDigest.Digest = ""
	bound, err := digestValue(withoutDigest)
	if err != nil || actual.Digest != bound {
		return nil, fmt.Errorf("receipt digest is not bound")
	}
	result := map[string]any{
		"schema": "gooo/meta-operation-provenance-verification/v1", "status": "VERIFIED", "source_digest": actual.SourceDigest,
		"receipt_digest": actual.Digest, "scenario_count": 4, "metric_count": 12, "fail_closed_count": 1, "direct_unknowns": 2, "dependency_blocks": 1,
	}
	result["digest"], err = digestValue(result)
	return result, err
}

func verifyScenario(actual scenario) error {
	if actual.Denominator != 12 || len(actual.Metrics) != 3 {
		return fmt.Errorf("scenario %q has a non-fixed metric set", actual.ID)
	}
	expected := map[string]struct{ edges, numerator int }{
		"complete": {edges: 12, numerator: 12}, "disconnected": {edges: 11, numerator: 11},
		"direct-unknown": {edges: 11, numerator: 11}, "dependency-blocked": {edges: 11, numerator: 11},
	}
	want, ok := expected[actual.ID]
	if !ok || actual.Graph.Nodes != 15 || actual.Graph.Edges != want.edges || actual.Numerator != want.numerator {
		return fmt.Errorf("scenario %q graph or fixed ratio is invalid", actual.ID)
	}
	wantKinds := map[string]int{"PRODUCES": 3, "CONSUMES": 3, "OPERATES": 3, "EVIDENCED_BY": 3}
	if actual.ID == "disconnected" {
		wantKinds["CONSUMES"] = 2
	}
	if actual.ID == "direct-unknown" || actual.ID == "dependency-blocked" {
		wantKinds["EVIDENCED_BY"] = 2
	}
	if !reflect.DeepEqual(actual.Graph.EdgeKinds, wantKinds) {
		return fmt.Errorf("complete scenario edge kinds are invalid")
	}
	claims := []struct{ id, family, claim string }{{"MOP-FOUNDATION-001", "FOUNDATION", "OPEN"}, {"MOP-COHERENCE-001", "COHERENCE", "DISCHARGED"}, {"MOP-REGRESSION-001", "REGRESSION", "REFUTED"}}
	wantDecisions := map[string]map[string]int{"complete": {"PASS": 3}, "disconnected": {"PASS": 2, "FAIL_CLOSED": 1}, "direct-unknown": {"PASS": 2, "UNKNOWN": 1}, "dependency-blocked": {"PASS": 1, "UNKNOWN": 2}}
	if !reflect.DeepEqual(actual.Decisions, wantDecisions[actual.ID]) {
		return fmt.Errorf("scenario %q decision counts are invalid", actual.ID)
	}
	for index, wantMetric := range claims {
		got := actual.Metrics[index]
		if got.ID != wantMetric.id || got.Family != wantMetric.family || got.Claim != wantMetric.claim || got.Denominator != 4 {
			return fmt.Errorf("scenario %q metric identity is invalid", actual.ID)
		}
		missing := ""
		if actual.ID == "disconnected" && got.ID == "MOP-COHERENCE-001" {
			missing = "consumer"
		}
		if (actual.ID == "direct-unknown" || actual.ID == "dependency-blocked") && got.ID == "MOP-FOUNDATION-001" {
			missing = "evidence"
		}
		wantNumerator, wantDecision := 4, "PASS"
		if missing != "" {
			wantNumerator = 3
		}
		if missing == "consumer" {
			wantDecision = "FAIL_CLOSED"
		}
		if missing == "evidence" {
			wantDecision = "UNKNOWN"
		}
		if actual.ID == "dependency-blocked" && got.ID == "MOP-REGRESSION-001" {
			wantDecision, missing = "UNKNOWN", "dependency"
		}
		if got.Numerator != wantNumerator || got.Decision != wantDecision || got.EvaluationState != "EVALUATED" {
			return fmt.Errorf("scenario %q metric decision is invalid", actual.ID)
		}
		if err := verifyLineage(got, missing); err != nil {
			return fmt.Errorf("scenario %q: %w", actual.ID, err)
		}
		if err := verifyIssue(got.Issue, actual.ID, got.ID, missing); err != nil {
			return err
		}
	}
	return nil
}

func verifyLineage(got metric, missing string) error {
	want := lineage{Producer: "producer:" + got.ID, Consumer: "consumer:" + got.ID, MetaOperation: "operation:" + got.ID, EvidencePath: "evidence:" + got.ID}
	if missing == "consumer" {
		want.Consumer = ""
	}
	if missing == "evidence" {
		want.EvidencePath = ""
	}
	if missing == "dependency" {
		want = lineage{Producer: "producer:" + got.ID, Consumer: "consumer:" + got.ID, MetaOperation: "operation:" + got.ID, EvidencePath: "evidence:" + got.ID}
	}
	if got.Lineage != want {
		return fmt.Errorf("metric %q lineage is invalid", got.ID)
	}
	return nil
}

func verifyIssue(got *issue, scenario, metricID, missing string) error {
	if missing == "" || missing == "dependency" {
		if missing == "" && got != nil {
			return fmt.Errorf("metric %q unexpectedly has an issue", metricID)
		}
		if missing == "dependency" && !reflect.DeepEqual(got, &issue{Stage: "DEPENDENCY", Step: "propagate-unknown", Reason: "UPSTREAM_UNKNOWN", Cause: "DEPENDENCY_BLOCK", BlockedBy: []string{"MOP-FOUNDATION-001"}}) {
			return fmt.Errorf("scenario %q dependency block is not explicit", scenario)
		}
		return nil
	}
	want := &issue{Stage: "LINEAGE", Step: "connect-consumer", Reason: "DISCONNECTED_METRIC", Cause: "DIRECT_CAUSE"}
	if missing == "evidence" {
		want = &issue{Stage: "BINDING", Step: "evidence-path", Reason: "REQUIRED_EVIDENCE_MISSING", Cause: "DIRECT_CAUSE"}
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("scenario %q metric %q issue is not independently classified", scenario, metricID)
	}
	return nil
}

func decodeExact(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}
func digest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func digestValue(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digest(payload), nil
}
