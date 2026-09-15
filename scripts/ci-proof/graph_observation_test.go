package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func graphObservationFixture() domainCommand {
	return domainCommand{
		Command: "go run ./cmd/gooo graph dump examples/billing/main.gooo",
		Status:  "deferred", Reason: "GRAPH_OBSERVER_NOT_RUN", OutputSHA256: digestBytes(nil),
		Observation: &graphObservation{
			State: "UNKNOWN", Stage: "DOMAIN_EVIDENCE", Step: "GRAPH_DUMP", Reason: "GRAPH_OBSERVER_NOT_RUN",
			UnknownClass: "DIRECT_MISSING", NextOperation: "RUN_GRAPH_DUMP_OBSERVER", BlockedBy: []string{},
		},
	}
}

func TestGraphObserverCoveragePreservesUnknown(t *testing.T) {
	graph := graphObservationFixture()
	if err := validateGraphObservation(graph); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	var restored domainCommand
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(graph, restored) {
		t.Fatal("UNKNOWN observer fields were not preserved")
	}
}

func TestGraphObserverCoverageRejectsUnsupportedClaims(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*domainCommand)
	}{
		{"state", func(g *domainCommand) { g.Observation.State = "FIXED_POINT" }},
		{"stage", func(g *domainCommand) { g.Observation.Stage = "" }},
		{"step", func(g *domainCommand) { g.Observation.Step = "" }},
		{"reason", func(g *domainCommand) { g.Observation.Reason = "NOT_IMPLEMENTED" }},
		{"envelope reason", func(g *domainCommand) { g.Reason = "NOT_IMPLEMENTED" }},
		{"class", func(g *domainCommand) { g.Observation.UnknownClass = "DEPENDENCY_BLOCKED" }},
		{"next operation", func(g *domainCommand) { g.Observation.NextOperation = "" }},
		{"missing frontier", func(g *domainCommand) { g.Observation.BlockedBy = nil }},
		{"invented blocker", func(g *domainCommand) { g.Observation.BlockedBy = []string{"language"} }},
		{"stale command", func(g *domainCommand) { g.Command = "go run ./cmd/gooo graph-dump examples/billing/main.gooo" }},
		{"verified", func(g *domainCommand) { g.Status = "verified" }},
		{"available", func(g *domainCommand) { g.Available = true }},
		{"output", func(g *domainCommand) { g.Output = "success" }},
		{"output digest", func(g *domainCommand) { g.OutputSHA256 = digestBytes([]byte("success")) }},
		{"missing observation", func(g *domainCommand) { g.Observation = nil }},
		{"stripped current reason", func(g *domainCommand) {
			g.Observation = nil
			g.Reason = ""
		}},
		{"new reason on legacy command", func(g *domainCommand) {
			g.Observation = nil
			g.Command = "go run ./cmd/gooo graph-dump examples/billing/main.gooo"
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			graph := graphObservationFixture()
			test.mutate(&graph)
			if err := validateGraphObservation(graph); err == nil {
				t.Fatal("unsupported observer claim was accepted")
			}
		})
	}
}

func TestHistoricalGraphObserverEnvelopeHasNoNewFields(t *testing.T) {
	graph := validProof().DomainEvidence.Graph
	data, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "observation") {
		t.Fatal("historical receipt bytes gained an unobserved coverage claim")
	}
	if err := validateGraphObservation(graph); err != nil {
		t.Fatal(err)
	}
}
