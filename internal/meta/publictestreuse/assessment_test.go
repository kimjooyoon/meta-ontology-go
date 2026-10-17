package publictestreuse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestGoooReusePolicyLiteralExpectations(t *testing.T) {
	policy, _ := loadReusePolicyFixture(t)
	data, err := os.ReadFile("testdata/reuse_decisions.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Cases []struct {
			ID       string                     `json:"id"`
			Inputs   map[string]json.RawMessage `json:"inputs"`
			Expected map[string]json.RawMessage `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &suite); err != nil || len(suite.Cases) != 17 {
		t.Fatal("seventeen independently authored expectations are required", err)
	}
	for _, row := range suite.Cases {
		t.Run(row.ID, func(t *testing.T) {
			if len(row.Inputs) != 1 || len(row.Expected) != 1 {
				t.Fatal("case must identify exactly one activity")
			}
			for name, input := range row.Inputs {
				id := "gooo://test-reuse/assess-replay/v1"
				if name == "AssessOriginal" {
					id = "gooo://test-reuse/assess-original/v1"
				} else if name != "AssessReplay" {
					t.Fatal("unknown fixture activity")
				}
				var expected Assessment
				if json.Unmarshal(row.Expected[name], &expected) != nil {
					t.Fatal("missing literal expectation")
				}
				value, actual, err := policy.assess(context.Background(), name, id, input)
				if err != nil || actual != expected || value.SourceSHA256 != policy.DecisionSourceDigest {
					t.Fatalf("Gooo decision differs: got %+v want %+v: %v", actual, expected, err)
				}
			}
		})
	}
}

func TestGoooReusePolicySourceOwnsInputProfileDecision(t *testing.T) {
	policy, source := loadReusePolicyFixture(t)
	observed := ReuseObservation{Requested: true, ReceiptState: "VALID", PolicyState: "MATCH",
		BindingState: "MATCH", InputProfile: ExecutionPackageProfile}
	first, a, err := policy.AssessReplay(context.Background(), observed)
	if err != nil || a.Decision != DecisionClosed || !a.Reuse {
		t.Fatal("original source did not classify the exact profile", err)
	}
	changed := bytes.Replace(source, []byte(ExecutionPackageProfile), []byte("gooo/unobserved-input-profile/v1"), 1)
	if bytes.Equal(source, changed) {
		t.Fatal("source profile condition was not found")
	}
	restricted := loadChangedReusePolicy(t, changed)
	second, b, err := restricted.AssessReplay(context.Background(), observed)
	if err != nil || b.Decision != DecisionUnknown || b.Reuse || b.BlockedBy != "execution_package_inputs" ||
		first.SourceSHA256 == second.SourceSHA256 || first.ActivityID != second.ActivityID ||
		policy.SourceDigest != restricted.SourceDigest {
		t.Fatal("source-only condition did not change the observed decision and provenance", err)
	}
}

func TestGoooReusePolicyCancellationAndMalformedBody(t *testing.T) {
	policy, source := loadReusePolicyFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := policy.AssessReplay(ctx, ReuseObservation{}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled policy lost the cancellation", err)
	}
	if _, _, err := policy.AssessReplay(nil, ReuseObservation{}); err == nil {
		t.Fatal("nil context was accepted")
	}
	changed := bytes.Replace(source, []byte("return Outcome("), []byte("return UnavailablePolicyActivity("), 1)
	malformed := loadChangedReusePolicy(t, changed)
	if _, _, err := malformed.AssessReplay(context.Background(), ReuseObservation{}); err == nil {
		t.Fatal("unknown activity was treated as a completed decision")
	}
}

func loadReusePolicyFixture(t *testing.T) (Policy, []byte) {
	t.Helper()
	return loadChangedReusePolicy(t, decisionSource), append([]byte(nil), decisionSource...)
}

func loadChangedReusePolicy(t *testing.T, program []byte) Policy {
	t.Helper()
	filename := "../../../examples/self-improvement-discovery/project.gooo"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := loadWithDecisionSource(filename, source, program)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
