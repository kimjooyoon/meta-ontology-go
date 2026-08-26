package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func completeInput() Input {
	entity := "urn:selectiveci:entity/order"
	obligation := "urn:selectiveci:obligation/order"
	command := "urn:selectiveci:command/test"
	guard := "urn:selectiveci:command/guard"
	base := SnapshotManifest{SchemaVersion: ManifestSchemaVersion, Files: []SnapshotFile{{Path: "billing/order.gooo", BlobDigest: digest("base"), SemanticIDs: []string{entity}}}}
	head := SnapshotManifest{SchemaVersion: ManifestSchemaVersion, Files: []SnapshotFile{{Path: "billing/order.gooo", BlobDigest: digest("head"), SemanticIDs: []string{entity}}}}
	base.Digest, head.Digest = base.ComputedDigest(), head.ComputedDigest()
	registry := Registry{SchemaVersion: RegistrySchemaVersion, PolicyDigest: digest("policy"), Nodes: []impactgraph.Node{{ID: entity, Kind: impactgraph.NodeKindSemantic}, {ID: obligation, Kind: impactgraph.NodeKindObligation}}, DependencyEdges: []DependencyEdge{}, Obligations: []ObligationBinding{{ID: obligation, Subject: entity, CommandIDs: []string{command}}}, Commands: []Command{{ID: command, Argv: []string{"go", "test"}, WorkingDir: ".", CPUWorkUnits: 100, MemoryBytes: 1000}}, GlobalGuardCommands: []Command{{ID: guard, Argv: []string{"gofmt", "-l", "."}, WorkingDir: ".", CPUWorkUnits: 50, MemoryBytes: 1000}}}
	registry.Digest = registry.ComputedDigest()
	return Input{SchemaVersion: SchemaVersion, Base: base, Head: head, Registry: registry, CPUCapacity: 1000, Receipts: []Receipt{receipt(command, head.Digest, command), receipt(guard, head.Digest, guard)}, ProvenancePaths: []ProvenancePath{provenance(command), provenance(guard)}}
}
func refreshInputDigests(input *Input) {
	input.Base.Digest = input.Base.ComputedDigest()
	input.Head.Digest = input.Head.ComputedDigest()
	input.Registry.Digest = input.Registry.ComputedDigest()
}
func receipt(commandID, snapshot string, label string) Receipt {
	return Receipt{CommandID: commandID, SnapshotDigest: snapshot, Envelope: resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: digest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: 100, PeakRSSBytes: 1000, ReadBytes: 1000, WriteBytes: 1000}, Samples: []resourceenvelope.Sample{{CPUCoreNS: 1, WallNS: 10}, {CPUCoreNS: 10, WallNS: 10}, {CPUCoreNS: 20, WallNS: 10}, {CPUCoreNS: 30, WallNS: 10}, {CPUCoreNS: 40, WallNS: 10}, {CPUCoreNS: 50, WallNS: 10}}}}
}
func provenance(commandID string) ProvenancePath {
	subject := semantic.MustIdentity("urn:selectiveci:path/" + commandID)
	object := semantic.MustIdentity("urn:selectiveci:path/result/" + commandID)
	record := semantic.MustIdentity("urn:selectiveci:record/" + commandID)
	evidenceID := semantic.MustIdentity("urn:selectiveci:evidence/" + commandID)
	ruleID := semantic.MustIdentity("urn:selectiveci:rule/v1")
	before := semantic.SnapshotDigests{Source: digest("before-" + commandID)}
	after := semantic.SnapshotDigests{Source: digest("after-" + commandID)}
	controls := semantic.InferenceControls{}
	edge := semantic.InferenceEdge{RecordID: record, SubjectID: subject, ObjectID: object, Rule: semantic.RuleBinding{ID: ruleID, Version: "1", Digest: digest("rule")}, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDeclaration, Ordinal: 1}, Before: before, After: after, Authority: semantic.AuthorityBinding{Layer: semantic.AuthoritySource, Effect: semantic.AuthorityDeclare}, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: digest("evidence-" + commandID)}}, Controls: controls, Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{semantic.MustIdentity("urn:selectiveci:source/" + commandID)}}
	evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: digest("evidence-" + commandID), Before: before, After: after, Controls: controls, SourceBacked: true}
	return ProvenancePath{CommandID: commandID, Path: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: []semantic.InferenceEdge{edge}, Evidence: []semantic.InferenceEvidence{evidence}}, Requirement: PathRequirement{PathID: semantic.MustIdentity("urn:selectiveci:path-id/" + commandID).String(), RecordIDs: []string{record.String()}, ExpectedKinds: []string{string(semantic.InferenceAuthoritativeDeclaration)}, StartID: subject.String(), EndID: object.String()}}
}
func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func containsString(value, needle string) bool {
	return strings.Contains(value, needle)
}
