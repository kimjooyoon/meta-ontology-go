package couplingexplain

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type fixtureAdapter struct{}

func (fixtureAdapter) DecodeVerifiedEnvelope(data []byte) (VerifiedEnvelope, error) {
	return DecodeVerifiedEnvelope(data)
}

type detectorFixtureAdapter struct {
	envelope VerifiedEnvelope
}

func (adapter detectorFixtureAdapter) AdaptDetectorSnapshot(DetectorSnapshot) (VerifiedEnvelope, error) {
	return adapter.envelope, nil
}

func fixtureEnvelope(t *testing.T, claim ChangeClaim, verdict EnvelopeVerdict) (Request, VerifiedEnvelope) {
	t.Helper()
	bindingDigest := digest("binding")
	termDigest := "9efc05819657dc7d15c80fbbfc904f59e45f8ff9b4df78be0a6677f0f75598e2"
	evidenceDigest := digest("evidence")
	control := Control{RequestVersion: 7, ObservedVersion: 7, RequestCancellationVersion: 11, ObservedCancellationVersion: 11}
	binding := SnapshotBinding{SnapshotDigest: digest("snapshot"), RegistryDigest: digest("registry"), SourceMapDigest: digest("source-map"), ManifestDigest: digest("manifest"), ToolchainDigest: digest("toolchain"), ProfileDigest: digest("profile"), DetectorInputDigest: digest("detector-input"), DetectorResultDigest: digest("detector-result"), VerifierResultDigest: digest("verifier-result"), Control: control}
	path := PathSummary{PathID: "path://pay-order", StartID: "code://billing/pay-order", EndID: "evidence://coupling", StepCount: 3, PathDigest: "289791d1fab43470f301f2a023e1d6dad2a003a8077ddb71ebde9864a744a2dc", Steps: []PathStep{
		{FromID: "code://billing/pay-order", ToID: "owner://billing/pay-order", Kind: semantic.InferenceDerivedProjection, Phase: semantic.PhasePlacement{Phase: semantic.PhaseProjection, Ordinal: 1}, InputDigest: bindingDigest, OutputDigest: digest("owner")},
		{FromID: "owner://billing/pay-order", ToID: "term://pay-order", Kind: semantic.InferenceAuthoritativeDeclaration, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDeclaration, Ordinal: 2}, InputDigest: digest("owner"), OutputDigest: termDigest},
		{FromID: "term://pay-order", ToID: "evidence://coupling", Kind: semantic.InferenceIndependentVerification, Phase: semantic.PhasePlacement{Phase: semantic.PhaseVerification, Ordinal: 3}, InputDigest: termDigest, OutputDigest: evidenceDigest, EvidenceRef: "evidence://coupling"},
	}}
	receipt := ReceiptSummary{ReceiptID: "receipt://pay-order", SurfaceID: "surface://pay-order", ChangeClaim: claim, ReceiptKind: semantic.SemanticDelta, BeforeIRDigest: digest("before"), AfterIRDigest: digest("after"), CanonicalDelta: "owner=billing://pay-order relation=used", DeltaDigest: "f083f2025689f6b547605d267145bd5469caed9e88a3a54fc2de5af495fda744", ReceiptDigest: "67389bf778c30affeb56f6c691c889711e3c8f4d280a1cc7f971aae544865757", OriginPathID: path.PathID, EvidenceRefs: []string{"evidence://coupling"}}
	if claim == ClaimNoDelta {
		receipt.ReceiptKind = semantic.NoSemanticDelta
		receipt.AfterIRDigest = receipt.BeforeIRDigest
		receipt.CanonicalDelta = ""
		receipt.DeltaDigest = ""
		receipt.ReceiptDigest = "ddabc5991bad8d2b94a1fe103972c080c51064ed38f84ac5f6796ca050ec4288"
	}
	verifier := VerifierSummary{EvidenceID: "evidence://coupling", ReceiptID: receipt.ReceiptID, State: VerifierPass, Independent: true, EvidenceDigest: evidenceDigest, VerifierDigest: "96254158ba10b70351e9959fa5782148e3e1990e49f9ec3d2174404b60b3065e", EvidenceRefs: []string{path.PathID}}
	envelope := VerifiedEnvelope{Schema: "gooo-coupling-explanation/v1", Binding: binding, CodeBinding: CodeBindingSummary{CodeSymbolID: "code://billing/pay-order", SemanticOwnerID: "owner://billing/pay-order", RegisteredSurfaceID: receipt.SurfaceID, SourceMapID: "sourcemap://pay-order", BindingDigest: bindingDigest, CodeBindingDigest: "16d0d29bf279a96f342aeff1e936a367889ae1fa54d83c376734fe679d148907"}, SemanticOwner: "owner://billing/pay-order", Term: TermSummary{TermID: "term://pay-order", SemanticOwnerID: "owner://billing/pay-order", Version: "v1", DefinitionDigest: termDigest}, OriginPath: path, Receipt: receipt, Verifier: verifier, Verdict: verdict, EvidenceDigest: evidenceDigest}
	if verdict != VerdictVerified {
		envelope.NoLinkReason = ReasonMissing
		envelope.Diagnostics = []Diagnostic{{Code: "fixture-no-link"}}
	}
	envelope.EnvelopeDigest = fixtureEnvelopeDigest(claim, verdict)
	request := Request{CodeSymbolID: envelope.CodeBinding.CodeSymbolID, SnapshotDigest: binding.SnapshotDigest, RegistryDigest: binding.RegistryDigest, SourceMapDigest: binding.SourceMapDigest, ManifestDigest: binding.ManifestDigest, ToolchainDigest: binding.ToolchainDigest, ProfileDigest: binding.ProfileDigest, DetectorInputDigest: binding.DetectorInputDigest, DetectorResultDigest: binding.DetectorResultDigest, VerifierResultDigest: binding.VerifierResultDigest, EnvelopeDigest: envelope.EnvelopeDigest, Control: control}
	return request, envelope
}

func fixtureEnvelopeDigest(claim ChangeClaim, verdict EnvelopeVerdict) string {
	if claim == ClaimDelta {
		switch verdict {
		case VerdictVerified:
			return "114cf323cd5eaabe2834ce8277fb04ed275da3db5054fe27f5b98ea1e7e4b33f"
		case VerdictUnknown:
			return "39196e068070184f8b621cd83e0498790af4c72e69f90dc1a9bc678404eb545f"
		default:
			return "421d65fe27e0313c9cb73611ab0bc85a3d43b8309436cd1854eb4bca3706d47e"
		}
	}
	switch verdict {
	case VerdictVerified:
		return "7d1772935ff8626833ab57580387ebacddbf65696bbcdb698bd60861fc25c19b"
	case VerdictUnknown:
		return "ea58af7971a87d590c191f85fcfc71dc530aa0d1f03dc388080217f8df3d46b1"
	default:
		return "728cae826ff969b3450ae9bba7d3452d7f59827eeb2988f015ae96570d9dc25a"
	}
}

func refreshEnvelopeDigest(request *Request, envelope *VerifiedEnvelope) {
	envelope.EnvelopeDigest = envelope.Digest()
	request.EnvelopeDigest = envelope.EnvelopeDigest
}

func digest(value string) string { return DigestBytes([]byte(value)) }

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestNoLinkReasonsAreClosed(t *testing.T) {
	for _, value := range []LinkReason{ReasonAmbiguous, ReasonStale, ReasonUnregistered, ReasonMissing, ReasonNotVerified} {
		if !validReason(value) {
			t.Fatal(value)
		}
	}
}
