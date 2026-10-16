package metricprogram_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
)

func TestCompileRequestBytesOriginFresh1731(t *testing.T) {
	for _, mode := range []string{"original", "strategy-format", "verification-format"} {
		t.Run(mode, func(t *testing.T) {
			payload, verification := fixturePayloads(t)
			var plan metricprogram.StrategyPlan
			if err := json.Unmarshal(payload, &plan); err != nil {
				t.Fatal(err)
			}
			plan.Bindings[0].MetaOperation = "unregistered://request/origin"
			payload = fixtureJSON(t, plan)
			var formatted bytes.Buffer
			if mode == "strategy-format" {
				if err := json.Indent(&formatted, payload, "", "  "); err != nil {
					t.Fatal(err)
				}
				payload = formatted.Bytes()
			}
			if mode == "verification-format" {
				if err := json.Indent(&formatted, verification, "", "  "); err != nil {
					t.Fatal(err)
				}
				verification = formatted.Bytes()
			}
			_, _, err := metricprogram.Compile(payload, verification)
			var refusal *unresolvedpolicy.OperationRefusal
			if !errors.As(err, &refusal) {
				t.Fatal(err)
			}
			raw := fixtureJSON(t, refusal)
			var value struct {
				Origin struct {
					Inputs map[string]string `json:"inputs"`
				} `json:"origin"`
			}
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			if value.Origin.Inputs["strategy_payload_digest"] != fixtureDigest(string(payload)) ||
				value.Origin.Inputs["verification_payload_digest"] != fixtureDigest(string(verification)) {
				t.Fatalf("refusal did not retain actual request byte identity: %s", raw)
			}
		})
	}
}
