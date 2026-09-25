package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"testing"
)

func TestObligationCoverageOneRootTwoObligationsAndTransitiveDependency(t *testing.T) {
	t.Run("two obligations", func(t *testing.T) {
		input := completeInput()
		obligation := "urn:selectiveci:obligation/order-2"
		command := "urn:selectiveci:command/test-2"
		input.Registry.Nodes = append(input.Registry.Nodes, impactgraph.Node{ID: obligation, Kind: impactgraph.NodeKindObligation})
		input.Registry.Obligations = append(input.Registry.Obligations, ObligationBinding{ID: obligation, Subject: "urn:selectiveci:entity/order", CommandIDs: []string{command}})
		input.Registry.Commands = append(input.Registry.Commands, Command{ID: command, Argv: []string{"go", "test"}, WorkingDir: ".", CPUWorkUnits: 100, MemoryBytes: 1000})
		input.Registry.Digest = input.Registry.ComputedDigest()
		coverageInput := coverageInputFromPlanInput(t, input, "urn:selectiveci:entity/order")
		got := ObserveObligationCoverage(coverageInput)
		if got.Decision != CoverageDecisionExact || got.RequiredObligationCount != 2 || got.BoundCommandCount != 2 || got.DeterministicWorkUnits != 5 {
			t.Fatalf("two-obligation coverage = %#v", got)
		}
		coverageInput.Registry.Obligations[1].CommandIDs = []string{coverageInput.Registry.Obligations[0].CommandIDs[0]}
		coverageInput.Registry.Digest = coverageInput.Registry.ComputedDigest()
		coverageInput.Graph.RegistryDigest = coverageInput.Registry.Digest
		got = ObserveObligationCoverage(coverageInput)
		if got.Decision != CoverageDecisionExact || got.BoundCommandCount != 1 || got.DeterministicWorkUnits != 5 {
			t.Fatalf("shared-command coverage = %#v", got)
		}
	})
	t.Run("typed transitive dependency", func(t *testing.T) {
		input := typedTransitiveCoverageInput(t)
		got := ObserveObligationCoverage(input)
		if got.Decision != CoverageDecisionExact || got.Reason != CoverageReasonComplete || got.RequiredObligationCount != 1 || got.BoundCommandCount != 1 {
			t.Fatalf("transitive coverage = %#v", got)
		}
	})
}
