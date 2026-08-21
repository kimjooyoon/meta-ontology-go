package main

import (
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
)

func (f *shadowFixture) reencodeAll() {
	f.files = map[string][]byte{}
	f.files["base_snapshot.json"], _ = f.base.CanonicalJSON()
	f.files["head_snapshot.json"], _ = f.head.CanonicalJSON()
	f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
	f.reencodeProof()
	f.reencodeLane()
}
func (f *shadowFixture) reencodeProof() {
	f.files["evidence_input.json"], _ = proofsci.EncodeInput(f.proofInput)
}
func (f *shadowFixture) reencodeLane() {
	f.files["lane_input.json"], _ = lanesci.EncodeInputJSON(f.laneInput)
}
func (f shadowFixture) reader() shadowMapReader { return shadowMapReader(f.files) }
func (f shadowFixture) args() []string {
	return []string{"shadow", "--base-snapshot", "base_snapshot.json", "--head-snapshot", "head_snapshot.json", "--plan-input", "plan_input.json", "--evidence-input", "evidence_input.json", "--lane-input", "lane_input.json"}
}
func (f shadowFixture) argsReversed() []string {
	return []string{"shadow", "--lane-input", "lane_input.json", "--evidence-input", "evidence_input.json", "--plan-input", "plan_input.json", "--head-snapshot", "head_snapshot.json", "--base-snapshot", "base_snapshot.json"}
}
func (f *shadowFixture) reverseInputs() {
	f.planInput.Registry.Commands = append([]plannersci.Command(nil), f.planInput.Registry.Commands...)
	if len(f.proofInput.InferencePath.Edges) > 1 {
		for left, right := 0, len(f.proofInput.InferencePath.Edges)-1; left < right; left, right = left+1, right-1 {
			f.proofInput.InferencePath.Edges[left], f.proofInput.InferencePath.Edges[right] = f.proofInput.InferencePath.Edges[right], f.proofInput.InferencePath.Edges[left]
		}
	}
	if len(f.proofInput.InferencePath.Evidence) > 1 {
		for left, right := 0, len(f.proofInput.InferencePath.Evidence)-1; left < right; left, right = left+1, right-1 {
			f.proofInput.InferencePath.Evidence[left], f.proofInput.InferencePath.Evidence[right] = f.proofInput.InferencePath.Evidence[right], f.proofInput.InferencePath.Evidence[left]
		}
	}
	f.laneInput.OwnedPathPrefixes = append([]string(nil), f.laneInput.OwnedPathPrefixes...)
	f.laneInput.ChangedPaths = append([]string(nil), f.laneInput.ChangedPaths...)
	f.reencodeAll()
}
