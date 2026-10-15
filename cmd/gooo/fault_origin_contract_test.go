package main

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestFaultOriginRecordedContractRejectsBrokenLinks(t *testing.T) {
	for _, mode := range []string{"missing-site", "duplicate-site", "site-content", "different-site", "kind", "divisor", "root", "actual", "passed", "blocked", "field-filter"} {
		runtime := savedFaultOriginRuntime(t, "public-outputs/public-local-division.json")
		query := bodyexecution.NativeOriginQuery{Kind: "faults", Limit: 1}
		delivery := &runtime.Traces[2].Deliveries[0]
		switch mode {
		case "missing-site":
			runtime.FaultSites = nil
		case "duplicate-site":
			runtime.FaultSites = append(runtime.FaultSites, runtime.FaultSites[0])
		case "site-content":
			runtime.FaultSites[0].Expression = "different"
		case "different-site":
			delivery.Fault.Site.Variable = "different"
		case "kind":
			delivery.Fault.Kind = "other"
		case "divisor":
			delivery.Fault.Right = 1
		case "root":
			delivery.ActivityID = "other"
		case "actual":
			delivery.Actual = json.RawMessage("0")
		case "passed":
			*delivery.Passed = true
		case "blocked":
			delivery.BlockedBy = []string{"other"}
		case "field-filter":
			query.FieldID = "unrecorded-field"
		}
		if _, _, err := bodyexecution.SelectRecordedNativeOrigins(runtime, query); err == nil {
			t.Fatal("accepted a broken origin link", mode)
		}
	}
}

func TestFaultOriginReadIsNotAProducerReexecution(t *testing.T) {
	runtime := savedFaultOriginRuntime(t, "public-outputs/public-local-division.json")
	runtime.Traces[2].Deliveries[0].Fault.Left = 1009
	rows, report, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "faults", Limit: 1})
	if err != nil || len(rows) != 1 || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("recorded operand selection", err)
	}
	var fact bodyexecution.CompositionFaultFact
	if err := json.Unmarshal(rows[0], &fact); err != nil || fact.Left != 1009 {
		t.Fatal("reader recomputed a recorded operand", err, fact)
	}
}
