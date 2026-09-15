package main

import (
	"bytes"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"testing"
)

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
