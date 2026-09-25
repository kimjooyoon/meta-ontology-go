package couplingexplain

import (
	"context"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"testing"
)

func TestUpstreamDecisionOrderingIsCanonical(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
	envelope.Upstream = &UpstreamEvidence{
		SourceMapDigest: request.SourceMapDigest, ManifestDigest: request.ManifestDigest,
		DetectorInputDigest: request.DetectorInputDigest, DetectorResultDigest: request.DetectorResultDigest,
		DetectorStatus:  detector.StatusUnknown,
		DetectorReasons: []detector.Reason{{Code: detector.ReasonStaleInput, Detail: "stale"}, {Code: detector.ReasonRequiredInputMissing, Detail: "missing"}},
	}
	refreshEnvelopeDigest(&request, &envelope)
	first := Explain(context.Background(), request, envelope)
	firstJSON, err := first.CanonicalJSON(ViewExpanded)
	if err != nil {
		t.Fatal(err)
	}
	envelope.Upstream.DetectorReasons[0], envelope.Upstream.DetectorReasons[1] = envelope.Upstream.DetectorReasons[1], envelope.Upstream.DetectorReasons[0]
	refreshEnvelopeDigest(&request, &envelope)
	second := Explain(context.Background(), request, envelope)
	secondJSON, err := second.CanonicalJSON(ViewExpanded)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) || first.EvidenceDigest != second.EvidenceDigest {
		t.Fatalf("upstream permutation changed canonical output: first=%s second=%s", firstJSON, secondJSON)
	}
}

type liveFixtureAdapter struct {
	binding SnapshotBinding
}

func literalDetectorResult() detector.Result {
	return detector.Result{
		Schema: detector.ResultSchemaV1, Status: detector.StatusUnknown,
		Reasons:           []detector.Reason{{Code: detector.ReasonAuthorityInputSelfBound, Detail: "evaluator authority context is missing"}},
		FullSuiteRequired: true,
		InputDigest:       "265a7627c123865b1cb0a3cadfc74b0d9e079cfa85a78dfbc1534368d73c2beb",
		Digest:            "ceadad25cf2b1fb7d3af40568caa658634eef34a0f7e7e7b75ac4253e04bfd65",
	}
}
func literalDetectorResultBytes() []byte {
	return []byte(`{"schema":"gooo/code-semantic-coupling-result/v1","status":"UNKNOWN","accepted_surface_ids":null,"reasons":[{"code":"COUPLING_AUTHORITY_INPUT_SELF_BOUND","detail":"evaluator authority context is missing"}],"observation":{"changed_surfaces":{"known":false,"value":0},"receipts":{"known":false,"value":0},"inference_records":{"known":false,"value":0},"inference_paths":{"known":false,"value":0},"deterministic_work":{"known":false,"value":0},"resource_work":{"known":false,"value":0},"cpu":{"known":false,"value":0},"memory":{"known":false,"value":0}},"full_suite_required":true,"input_digest":"265a7627c123865b1cb0a3cadfc74b0d9e079cfa85a78dfbc1534368d73c2beb","digest":"ceadad25cf2b1fb7d3af40568caa658634eef34a0f7e7e7b75ac4253e04bfd65"}`)
}
func (adapter liveFixtureAdapter) AdaptLiveSnapshot(snapshot LiveSnapshot) (VerifiedEnvelope, error) {
	return MissingLiveLinkEnvelope(adapter.binding, snapshot)
}
