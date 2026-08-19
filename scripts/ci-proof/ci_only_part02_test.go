package main

import (
	"encoding/json"
	"testing"
)

func TestCIBranchProtectionRequiresExactGitHubAppBindings(t *testing.T) {
	bundle := validProof()
	bundle.BaseRef = "main"
	mainSnapshot := validBranchProtection(bundle)
	mainSnapshot.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	mainSnapshot.RequiredCheckBindings = requiredCheckBindingsFor(mainSnapshot.RequiredChecks)
	mainSnapshot.Digest = digestBranchProtection(mainSnapshot)
	mutations := []func(*branchProtection){
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].AppID = 1 },
		func(snapshot *branchProtection) { snapshot.RequiredCheckBindings[0].Context = "other" },
		func(snapshot *branchProtection) {
			snapshot.RequiredCheckBindings[1].Context = snapshot.RequiredCheckBindings[0].Context
		},
	}
	for index, mutate := range mutations {
		snapshot := mainSnapshot
		snapshot.RequiredCheckBindings = append([]requiredCheckBinding(nil), mainSnapshot.RequiredCheckBindings...)
		mutate(&snapshot)
		if branchProtectionReadyFor(snapshot, "main") {
			t.Fatalf("invalid app binding mutation %d was accepted", index)
		}
	}
}
func TestCIBranchProtectionBindingDigestRoundTripsStructuredChecks(t *testing.T) {
	bundle := validProof()
	bundle.BaseRef = "main"
	protection := validBranchProtection(bundle)
	protection.RequiredChecks = append(append([]string(nil), proofJobs...), "CI guardian")
	protection.RequiredCheckBindings = requiredCheckBindingsFor(protection.RequiredChecks)
	protection.Digest = digestBranchProtection(protection)
	data, err := json.Marshal(protection)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip branchProtection
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if len(roundTrip.RequiredCheckBindings) != len(proofJobs)+1 || digestBranchProtection(roundTrip) != protection.Digest {
		t.Fatalf("structured app bindings were not preserved in the protection digest: %+v", roundTrip.RequiredCheckBindings)
	}
}
