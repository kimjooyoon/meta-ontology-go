package publicresolutionrepair

import (
	"encoding/json"
	"testing"
)

func TestCounterexampleRetainsObservedReportSchema(t *testing.T) {
	policy := Policy{Partitions: []Partition{{ID: "inventory", TestName: "TestInventoryPartition"}}}
	aggregate, _ := json.Marshal(v15Aggregate{Decision: DecisionClosed, Policy: v15Policy{SourceDigest: "fixture",
		Cases: []v15PolicyCase{{ID: OriginalCounterexampleCaseID, Decision: DecisionRefuted,
			Changed: "shared-contract", GraphVariant: "missing-edge", Option: "omitted-target=inventory"}}}})
	for _, schema := range []string{"gooo/public-partial-test-reuse-report/v1", "gooo/public-partial-test-reuse-report/v2", "", "unknown"} {
		t.Run(schema, func(t *testing.T) {
			raw, _ := json.Marshal(v15Case{Schema: schema, CaseID: OriginalCounterexampleCaseID,
				Decision: DecisionRefuted, Reason: "FAIL_CLOSED_PARTIAL_REUSE_CONTRADICTION"})
			value, err := LoadCounterexample(raw, aggregate, policy)
			if schema == "" || schema == "unknown" {
				if err == nil {
					t.Fatal("unknown source schema was invented", value)
				}
			} else if err != nil || value.RecordSchema != schema {
				t.Fatal("counterexample lost its observed source schema", value, err)
			}
		})
	}
}
