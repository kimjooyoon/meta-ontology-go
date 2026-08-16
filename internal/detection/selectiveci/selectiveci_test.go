package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestPlanContractTable(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Input)
		status Status
		reason string
	}{
		{name: "exact complete input", status: StatusSelective},
		{name: "unsupported schema", mutate: func(input *Input) { input.SchemaVersion = "gooo/selective-ci/v0" }, status: StatusFullSuiteFallback, reason: ReasonUnsupportedSchema},
		{name: "mismatched head digest", mutate: func(input *Input) { input.Head.Digest = digest("wrong") }, status: StatusFullSuiteFallback, reason: ReasonMismatchedDigest},
		{name: "invalid argv", mutate: func(input *Input) { input.Registry.Commands[0].Argv = nil }, status: StatusFullSuiteFallback, reason: ReasonInvalidArgv},
		{name: "resource arithmetic", mutate: func(input *Input) { input.Receipts[0].Envelope.Samples[3].WallNS = 0 }, status: StatusFullSuiteFallback, reason: ReasonResourceArithmetic},
		{name: "missing provenance", mutate: func(input *Input) { input.ProvenancePaths = input.ProvenancePaths[:1] }, status: StatusFullSuiteFallback, reason: ReasonAmbiguousPath},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := completeInput()
			if test.mutate != nil {
				test.mutate(&input)
				if test.name == "invalid argv" {
					input.Registry.Digest = input.Registry.ComputedDigest()
				}
			}
			got := Plan(input)
			if got.Status != test.status || got.ReasonCode != test.reason {
				t.Fatalf("status/reason = %s/%s, want %s/%s", got.Status, got.ReasonCode, test.status, test.reason)
			}
			if got.CanonicalDigest == "" || got.CanonicalDigest != got.StableDigest() || got.Digest != got.CanonicalDigest {
				t.Fatalf("result is not sealed: %#v", got)
			}
			if got.Status == StatusFullSuiteFallback && len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs) != 0 {
				t.Fatalf("fallback retained partial selection: %#v", got)
			}
		})
	}
}

func TestManifestDeletionUsesStableIDUnion(t *testing.T) {
	input := completeInput()
	input.Head.Files = []SnapshotFile{}
	input.Head.Digest = input.Head.ComputedDigest()
	for i := range input.Receipts {
		input.Receipts[i].SnapshotDigest = input.Head.Digest
	}
	got := Plan(input)
	if got.Status != StatusSelective {
		t.Fatalf("deletion status = %s/%s", got.Status, got.ReasonCode)
	}
	if !reflect.DeepEqual(got.ChangedSemanticIDs, []string{"urn:selectiveci:entity/order"}) {
		t.Fatalf("changed IDs = %#v", got.ChangedSemanticIDs)
	}
}

func TestCanonicalPermutationIsByteIdentical(t *testing.T) {
	left := completeInput()
	right := completeInput()
	right.Base.Files[0].SemanticIDs = []string{"urn:selectiveci:entity/order"}
	right.Head.Files[0].SemanticIDs = []string{"urn:selectiveci:entity/order"}
	right.Receipts[0], right.Receipts[1] = right.Receipts[1], right.Receipts[0]
	right.ProvenancePaths[0], right.ProvenancePaths[1] = right.ProvenancePaths[1], right.ProvenancePaths[0]
	leftBytes, err := EncodeJSON(left)
	if err != nil {
		t.Fatal(err)
	}
	rightBytes, err := EncodeJSON(right)
	if err != nil {
		t.Fatal(err)
	}
	if string(leftBytes) != string(rightBytes) {
		t.Fatalf("permutation changed canonical input:\n%s\n%s", leftBytes, rightBytes)
	}
	if Plan(left).Canonical() != Plan(right).Canonical() {
		t.Fatal("permutation changed canonical output")
	}
}

func TestStrictJSONRejectsDuplicateAndUnknownFields(t *testing.T) {
	if _, err := DecodeJSON([]byte(`{"schema_version":"gooo/selective-ci/v1","schema_version":"gooo/selective-ci/v1"}`)); err == nil {
		t.Fatal("duplicate field was accepted")
	}
	if _, err := DecodeJSON([]byte(`{"schema_version":"gooo/selective-ci/v1","unknown":1}`)); err == nil {
		t.Fatal("unknown field was accepted")
	}
}

func TestStrictJSONRoundTripAndSealedPlan(t *testing.T) {
	input := completeInput()
	encoded, err := EncodeJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeJSON(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if input.Canonical() != decoded.Canonical() {
		t.Fatalf("round trip changed canonical input")
	}
	planBytes, err := EncodePlanJSON(Plan(input))
	if err != nil {
		t.Fatal(err)
	}
	plan := PlanJSON(planBytes)
	if plan.Status != StatusFullSuiteFallback || plan.ReasonCode != ReasonInvalidInput {
		t.Fatalf("sealed plan should not be accepted as input: %#v", plan)
	}
	if !containsString(string(planBytes), `"canonical_digest":"`) {
		t.Fatalf("plan output omitted canonical digest: %s", planBytes)
	}
}

func FuzzPlanJSONNeverPanics(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, data []byte) {
		result := PlanJSON(data)
		if result.CanonicalDigest == "" || result.CanonicalDigest != result.StableDigest() {
			t.Fatalf("unsealed result for arbitrary input: %#v", result)
		}
		if result.Status != StatusFullSuiteFallback && result.Status != StatusSelective {
			t.Fatalf("unknown result status %q", result.Status)
		}
	})
}

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
	edge := semantic.InferenceEdge{InferenceRecord: semantic.InferenceRecord{RecordID: record, SubjectID: subject, ObjectID: object, Rule: semantic.RuleBinding{ID: ruleID, Version: "1", Digest: digest("rule")}, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDeclaration, Ordinal: 1}, Before: before, After: after, Authority: semantic.AuthorityBinding{Layer: semantic.AuthoritySource, Effect: semantic.AuthorityDeclare}, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: digest("evidence-" + commandID)}}, Controls: controls}, Kind: semantic.InferenceAuthoritativeDeclaration, SourceRoots: []semantic.ID{semantic.MustIdentity("urn:selectiveci:source/" + commandID)}}
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
