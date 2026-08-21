package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
