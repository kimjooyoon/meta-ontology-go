package shadow

import (
	"bytes"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func configureProductionProof(t *testing.T, fixture *productionFixture, name string) {
	t.Helper()
	switch name {
	case "plan-proof-source-binding-mismatch":
		fixture.proofInput.Snapshots.Base.Source = productionDigest("tampered-source")
	case "plan-proof-semantic-binding-mismatch":
		fixture.proofInput.Snapshots.Head.Semantic = productionDigest("tampered-semantic")
	case "plan-proof-plan-digest-mismatch":
		fixture.proofInput.PlanDigest = productionDigest("tampered-plan")
	case "plan-proof-changed-roots-mismatch":
		fixture.proofInput.ChangedRootIDs = []semantic.ID{productionID(fixture.otherID)}
	case "plan-proof-selected-union-mismatch":
		fixture.proofInput.SelectedCommandIDs = []semantic.ID{productionID(fixture.otherID)}
	case "proof-fail-closed":
		fixture.proofInput.CommandReceipts[0].Digest = productionDigest("wrong-receipt")
	case "proof-unknown":
		fixture.proofInput.InferencePath.Edges = fixture.proofInput.InferencePath.Edges[:2]
	}
	fixture.encode()
}
func configureProductionLane(fixture *productionFixture, name string) {
	switch name {
	case "lane-unknown":
		fixture.laneInput.BaseSHA = ""
	case "lane-ineligible":
		fixture.laneInput.ActiveLeaseCount = 1
	case "lane-digest-tamper":
		fixture.laneInput.AheadCount = -1
	}
	fixture.encode()
}
func configureProductionMalformed(fixture *productionFixture, name string) {
	fixture.encode()
	switch name {
	case "malformed-unknown-field":
		fixture.files["evidence_input.json"] = append(fixture.files["evidence_input.json"], ' ')
		fixture.files["evidence_input.json"] = bytes.Replace(fixture.files["evidence_input.json"], []byte("{"), []byte(`{"extra":true,`), 1)
	case "malformed-duplicate-key":
		fixture.files["plan_input.json"] = bytes.Replace(fixture.files["plan_input.json"], []byte(`"schema_version":`), []byte(`"schema_version":"gooo/selective-ci/v1","schema_version":`), 1)
	case "malformed-trailing-json":
		fixture.files["lane_input.json"] = append(fixture.files["lane_input.json"], []byte(` {}`)...)
	case "incomplete-required-field":
		fixture.files["base_snapshot.json"] = bytes.Replace(fixture.files["base_snapshot.json"], []byte(`,"digest":"`+fixture.base.Digest+`"`), nil, 1)
	}
}
func configureProductionPermutation(t *testing.T, fixture *productionFixture) {
	t.Helper()
	fixture.planInput.Registry.Commands = append([]plannersci.Command(nil), fixture.planInput.Registry.Commands...)
	for left, right := 0, len(fixture.proofInput.InferencePath.Edges)-1; left < right; left, right = left+1, right-1 {
		fixture.proofInput.InferencePath.Edges[left], fixture.proofInput.InferencePath.Edges[right] = fixture.proofInput.InferencePath.Edges[right], fixture.proofInput.InferencePath.Edges[left]
	}
	for left, right := 0, len(fixture.proofInput.InferencePath.Evidence)-1; left < right; left, right = left+1, right-1 {
		fixture.proofInput.InferencePath.Evidence[left], fixture.proofInput.InferencePath.Evidence[right] = fixture.proofInput.InferencePath.Evidence[right], fixture.proofInput.InferencePath.Evidence[left]
	}
	fixture.reencode(t)
}
