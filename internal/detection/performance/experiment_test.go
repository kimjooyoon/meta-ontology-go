package performance

import (
	"strings"
	"testing"
)

func TestExperimentPassesForMatchingHostMeasurements(t *testing.T) {
	hypothesis := parserHypothesis()
	comparison, err := CompareEvidence(hypothesis, parserEvidence(GoHosted, StatusVerified, 3, 1, 4), parserEvidence(GoooHosted, StatusVerified, 3, 1, 4))
	if err != nil {
		t.Fatalf("CompareEvidence() error = %v", err)
	}
	if comparison.Outcome != OutcomePass || comparison.OperationDelta != 0 || comparison.AllocationDelta != 0 {
		t.Fatalf("comparison = %#v, want pass with zero deltas", comparison)
	}
}

func TestMetricCounterexampleFailsHypothesis(t *testing.T) {
	hypothesis := parserHypothesis()
	comparison, err := CompareEvidence(hypothesis, parserEvidence(GoHosted, StatusVerified, 3, 1, 4), parserEvidence(GoooHosted, StatusVerified, 4, 1, 4))
	if err != nil {
		t.Fatalf("CompareEvidence() error = %v", err)
	}
	if comparison.Outcome != OutcomeFail {
		t.Fatalf("comparison outcome = %q, want fail", comparison.Outcome)
	}
	if !strings.Contains(comparison.Gooo.Reason, "operation count") {
		t.Fatalf("gooo reason = %q, want operation counterexample", comparison.Gooo.Reason)
	}
}

func TestPlannedGoooStageIsDeferred(t *testing.T) {
	hypothesis := parserHypothesis()
	comparison, err := CompareEvidence(hypothesis, parserEvidence(GoHosted, StatusVerified, 3, 1, 4), parserEvidence(GoooHosted, StatusPlanned, 0, 0, 0))
	if err != nil {
		t.Fatalf("CompareEvidence() error = %v", err)
	}
	if comparison.Outcome != OutcomeDeferred || comparison.Gooo.Outcome != OutcomeDeferred {
		t.Fatalf("comparison = %#v, want deferred", comparison)
	}
}

func TestFixtureDigestMismatchIsRejected(t *testing.T) {
	hypothesis := parserHypothesis()
	evidence := parserEvidence(GoHosted, StatusVerified, 3, 1, 4)
	evidence.InputDigest = DigestInput("different input")
	if _, err := Evaluate(hypothesis, evidence); err == nil {
		t.Fatal("fixture digest mismatch was accepted")
	}
}

func TestHostRoleMismatchIsRejected(t *testing.T) {
	hypothesis := parserHypothesis()
	goEvidence := parserEvidence(GoooHosted, StatusVerified, 3, 1, 4)
	goooEvidence := parserEvidence(GoHosted, StatusVerified, 3, 1, 4)
	if _, err := CompareEvidence(hypothesis, goEvidence, goooEvidence); err == nil {
		t.Fatal("swapped host evidence was accepted")
	}
}

func parserHypothesis() Hypothesis {
	fixture := NewFixture("fixture://parser/minimal-activity", "parser", "package billing\nnamespace billing\nentity Order id \"billing://entity/order\"\nactivity PayOrder(Order) -> Order\n")
	return Hypothesis{
		ID:                  "hypothesis://performance/parser-stable-counts",
		Statement:           "fixed parser fixture has stable operation and allocation counts across hosts",
		Contract:            StandardContracts()[0],
		Fixture:             fixture,
		Repetitions:         4,
		ExpectedOperations:  3,
		ExpectedAllocations: 1,
	}
}

func parserEvidence(host Host, status Status, operations, allocations, repetitions uint64) RunEvidence {
	hypothesis := parserHypothesis()
	return RunEvidence{
		Host:                    host,
		Status:                  status,
		FixtureID:               hypothesis.Fixture.ID,
		InputDigest:             hypothesis.Fixture.InputDigest,
		Repetitions:             repetitions,
		OperationsPerIteration:  operations,
		AllocationsPerIteration: allocations,
		Source:                  "deterministic fixture test",
	}
}
