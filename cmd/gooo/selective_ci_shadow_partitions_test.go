package main

import (
	"bytes"
	"encoding/json"
	"testing"

	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSelectiveCIShadowMalformedUnknownAndDuplicateJSONFallback(t *testing.T) {
	cases := []struct {
		name      string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "malformed", component: "evidence_input", mutate: func(f *shadowFixture) { f.files["evidence_input.json"] = []byte("{") }},
		{name: "unknown field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = append(f.files["evidence_input.json"], []byte(" ")...)
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("{"), []byte("{\"unknown\":true,"), 1)
		}},
		{name: "duplicate field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("\"schema\":"), []byte("\"schema\":\"gooo-selective-ci-evidence/v1\",\"schema\":"), 1)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != "INPUT" || output.Component != test.component || output.Reason == "" {
				t.Fatalf("fallback = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowBindingAndPlannerPrecedence(t *testing.T) {
	cases := []struct {
		name      string
		stage     string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "stale snapshot", stage: "INPUT", component: "base_snapshot", mutate: func(f *shadowFixture) {
			f.files["base_snapshot.json"] = bytes.Replace(f.files["base_snapshot.json"], []byte("sha256:"), []byte("sha256:0"), 1)
		}},
		{name: "manifest mismatch", stage: "SNAPSHOT_BINDING", component: "base_manifest", mutate: func(f *shadowFixture) {
			f.planInput.Base.Files[0].SemanticIDs = []string{f.entityID, f.otherID}
			f.planInput.Base.Digest = f.planInput.Base.ComputedDigest()
			f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
		}},
		{name: "registry mismatch", stage: "REGISTRY_BINDING", component: "base_snapshot", mutate: func(f *shadowFixture) {
			f.base = buildAnalyzerShadowSnapshot(t, f.sourceBase+"// base\n", f.entityID, prefixedShadowDigest("different-registry"))
			f.files["base_snapshot.json"], _ = f.base.CanonicalJSON()
		}},
		{name: "planner fallback", stage: "PLAN", component: "planner", mutate: func(f *shadowFixture) {
			f.planInput.CPUCapacity = 1
			f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
		}},
		{name: "lane registry precedence", stage: "REGISTRY_BINDING", component: "lane", mutate: func(f *shadowFixture) {
			f.laneInput.RegistryDigest = shadowDigest("wrong-lane-registry")
			f.laneInput.ActiveLeaseCount = 1
			f.planInput.CPUCapacity = 1
			f.proofInput.RegistryDigest = shadowDigest("wrong-proof-registry")
			f.reencodeAll()
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != test.stage || output.Component != test.component || output.ExecutionAuthorized || !output.ShadowOnly {
				t.Fatalf("fallback = %#v", output)
			}
			if len(output.SelectedCommands) != 0 || len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 0 {
				t.Fatalf("fallback exposed selection = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowLaneRegistryBindingFailsClosedWithoutExecution(t *testing.T) {
	fixture := newShadowFixture(t)
	fixture.laneInput.RegistryDigest = shadowDigest("wrong-lane-registry")
	fixture.reencodeLane()
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stderr = %q", code, stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != "REGISTRY_BINDING" || output.Component != "lane" || output.Reason != "REGISTRY_DIGEST_MISMATCH" {
		t.Fatalf("lane registry fallback = %#v", output)
	}
	if output.ExecutionAuthorized || !output.ShadowOnly || len(output.SelectedCommands) != 0 || len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 0 || len(output.ResourceReceipts) != 0 {
		t.Fatalf("lane registry fallback exposed execution = %#v", output)
	}
	if bytes.Contains(stdout.Bytes(), []byte("gooo-shadow-sentinel")) || bytes.Contains(stdout.Bytes(), []byte("never-run")) {
		t.Fatal("lane registry fallback exposed or ran sentinel argv")
	}
}

func TestSelectiveCIShadowProofBindingFailurePrecedence(t *testing.T) {
	cases := []struct {
		name      string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "registry", component: "registry_digest", mutate: func(f *shadowFixture) {
			f.proofInput.RegistryDigest = shadowDigest("wrong-proof-registry")
			f.reencodeProof()
		}},
		{name: "plan", component: "plan_digest", mutate: func(f *shadowFixture) { f.proofInput.PlanDigest = shadowDigest("wrong-proof-plan"); f.reencodeProof() }},
		{name: "changed roots", component: "changed_root_ids", mutate: func(f *shadowFixture) {
			f.proofInput.ChangedRootIDs = []semantic.ID{commandIDToID(f.otherID)}
			f.reencodeProof()
		}},
		{name: "selected commands", component: "selected_command_ids", mutate: func(f *shadowFixture) {
			f.proofInput.SelectedCommandIDs = []semantic.ID{commandIDToID(f.otherID)}
			f.reencodeProof()
		}},
		{name: "snapshots", component: "snapshots", mutate: func(f *shadowFixture) {
			f.proofInput.Snapshots.Head.Semantic = shadowDigest("wrong-proof-snapshot")
			f.reencodeProof()
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Stage != "PLAN_PROOF_BINDING" || output.Component != test.component || output.Status != "FULL_SUITE_FALLBACK" {
				t.Fatalf("proof binding fallback = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowProofAndLaneFailClosedPartitions(t *testing.T) {
	cases := []struct {
		name   string
		stage  string
		mutate func(*shadowFixture)
	}{
		{name: "proof fail closed", stage: "PROOF_FAIL_CLOSED", mutate: func(f *shadowFixture) {
			f.proofInput.CommandReceipts[0].Digest = shadowDigest("wrong-receipt")
			f.reencodeProof()
		}},
		{name: "proof unknown", stage: "PROOF_UNKNOWN", mutate: func(f *shadowFixture) {
			f.proofInput.InferencePath.Edges = f.proofInput.InferencePath.Edges[:2]
			f.reencodeProof()
		}},
		{name: "lane unknown", stage: "LANE_UNKNOWN", mutate: func(f *shadowFixture) { f.laneInput.BaseSHA = ""; f.reencodeLane() }},
		{name: "lane ineligible", stage: "LANE_INELIGIBLE", mutate: func(f *shadowFixture) { f.laneInput.ActiveLeaseCount = 1; f.reencodeLane() }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != test.stage || output.ExecutionAuthorized || !output.ShadowOnly {
				t.Fatalf("partition fallback = %#v", output)
			}
		})
	}
}

func runShadowFixture(t *testing.T, fixture shadowFixture) selectiveCIShadowOutput {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.CanonicalDigest == "" || output.CanonicalDigest != output.stableDigest() {
		t.Fatalf("invalid output digest: %#v", output)
	}
	return output
}
