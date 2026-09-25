package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"testing"
)

func coverageInputFromPlanInput(t *testing.T, input Input, roots ...string) ObligationCoverageInput {
	t.Helper()
	graph, err := buildGraph(input)
	if err != nil {
		t.Fatalf("build coverage graph: %v", err)
	}
	return ObligationCoverageInput{SchemaVersion: ObligationCoverageSchemaVersion, Graph: graph, Registry: input.Registry, SnapshotDigest: input.Head.Digest, ChangedRootIDs: roots}
}
func typedTransitiveCoverageInput(t *testing.T) ObligationCoverageInput {
	t.Helper()
	root := "urn:selectiveci:entity/root"
	packageID := "urn:selectiveci:package/pkg"
	obligation := "urn:selectiveci:obligation/root"
	command := "urn:selectiveci:command/root"
	snapshot := digest("typed-snapshot")
	registry := Registry{SchemaVersion: RegistrySchemaVersion, PolicyDigest: digest("typed-policy"), Nodes: []impactgraph.Node{{ID: root, Kind: impactgraph.NodeKindSemantic}, {ID: packageID, Kind: impactgraph.NodeKindGoPackage}, {ID: obligation, Kind: impactgraph.NodeKindObligation}}, DependencyEdges: []DependencyEdge{{From: root, To: packageID, Kind: impactgraph.EdgeKindProjectsTo}, {From: packageID, To: obligation, Kind: impactgraph.EdgeKindAffects}}, Obligations: []ObligationBinding{{ID: obligation, Subject: packageID, CommandIDs: []string{command}}}, Commands: []Command{{ID: command, Argv: []string{"go", "test"}, WorkingDir: ".", CPUWorkUnits: 1, MemoryBytes: 1}}, GlobalGuardCommands: []Command{}}
	registry.Digest = registry.ComputedDigest()
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: snapshot, RegistryDigest: registry.Digest, PolicyDigest: registry.PolicyDigest, Nodes: registry.Nodes, Edges: []impactgraph.Edge{{From: root, To: packageID, Kind: impactgraph.EdgeKindProjectsTo}, {From: packageID, To: obligation, Kind: impactgraph.EdgeKindAffects}}}
	return ObligationCoverageInput{SchemaVersion: ObligationCoverageSchemaVersion, Graph: graph, Registry: registry, SnapshotDigest: snapshot, ChangedRootIDs: []string{root}}
}
