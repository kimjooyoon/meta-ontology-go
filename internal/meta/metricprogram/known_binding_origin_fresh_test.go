package metricprogram_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func TestKnownBindingFailureScopeFresh1729(t *testing.T) {
	for _, mode := range []string{"status", "evidence", "proof-family"} {
		t.Run(mode, func(t *testing.T) {
			payload, verification := fixturePayloads(t)
			var plan metricprogram.StrategyPlan
			if err := json.Unmarshal(payload, &plan); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "status":
				plan.Bindings[0].Status = "UNSATISFIED"
			case "evidence":
				plan.Bindings[0].EvidenceDigest = ""
			case "proof-family":
				plan.Bindings[0].Family = "FOUNDATION"
			}
			program, source, err := metricprogram.Compile(fixtureJSON(t, plan), verification)
			var unregistered *unresolvedpolicy.OperationRefusal
			if err == nil || errors.As(err, &unregistered) || program.Schema != "" || source != nil {
				t.Fatalf("known binding failure was reclassified or executed: %+v, %v", program, err)
			}
		})
	}
}
