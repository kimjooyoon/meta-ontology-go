package operationprovenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ReceiptSchema = "gooo/meta-operation-provenance-receipt/v1"
	ReportSchema  = "gooo/meta-operation-provenance-verification/v1"
	Toolchain     = "go1.27.0"
)

type metricDefinition struct {
	ID, Family, Claim string
}

var metricDefinitions = []metricDefinition{
	{ID: "MOP-FOUNDATION-001", Family: "FOUNDATION", Claim: "OPEN"},
	{ID: "MOP-COHERENCE-001", Family: "COHERENCE", Claim: "DISCHARGED"},
	{ID: "MOP-REGRESSION-001", Family: "REGRESSION", Claim: "REFUTED"},
}

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

func CanonicalSource() []byte { return []byte(canonicalSource) }

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal digest subject: %w", err)
	}
	return digestBytes(payload), nil
}

func definitions() []metricDefinition { return append([]metricDefinition(nil), metricDefinitions...) }
