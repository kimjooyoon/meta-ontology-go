package valueexecution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func selfImprovementExecutionReceipt(origin provenance.OriginChainObservation, id string, phase ExecutionPhase) ExecutionOriginReceipt {
	return ObserveExecutionOrigin(origin, Execution{
		ExecutionDigest: digestValue(id),
		Phase:           phase,
	})
}

func TestObserveExecutedSelfImprovementBindsCompletedReceipts(t *testing.T) {
	beforeOrigin := selfImprovementOrigin("before-execution")
	afterOrigin := selfImprovementOrigin("after-execution")
	environment := selfImprovementEnvironment("source-v1")
	observation := ObserveExecutedSelfImprovement(
		beforeOrigin, afterOrigin, environment, environment,
		selfImprovementExecutionReceipt(beforeOrigin, "execution-before", ExecutionPhaseCompleted),
		selfImprovementExecutionReceipt(afterOrigin, "execution-after", ExecutionPhaseCompleted),
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusImproved || observation.Reason != "METRIC_IMPROVED" {
		t.Fatalf("observation=%#v, want completed improvement", observation)
	}
	if observation.CandidateDigest == "" || !validDigest(observation.Digest) || !observation.NonAuthorizing {
		t.Fatalf("observation must preserve detached evidence: %#v", observation)
	}
}

func TestObserveExecutedSelfImprovementDoesNotPromoteInFlightReceipt(t *testing.T) {
	beforeOrigin := selfImprovementOrigin("before-running")
	afterOrigin := selfImprovementOrigin("after-running")
	environment := selfImprovementEnvironment("source-v1")
	observation := ObserveExecutedSelfImprovement(
		beforeOrigin, afterOrigin, environment, environment,
		selfImprovementExecutionReceipt(beforeOrigin, "execution-before", ExecutionPhaseCompleted),
		selfImprovementExecutionReceipt(afterOrigin, "execution-after", ExecutionPhaseRunning),
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusUnknown || observation.Reason != "EXECUTION_NOT_COMPLETED" {
		t.Fatalf("observation=%#v, want UNKNOWN for in-flight execution", observation)
	}
}
