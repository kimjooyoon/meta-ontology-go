package generation

import (
	"bytes"
	"testing"
)

func TestCallbackPreviewContractIsTypedAndPreviewOnly(t *testing.T) {
	contract, err := LoadCallbackPreviewContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract.SourceDigest == "" || contract.SemanticDigest == "" || contract.InputEntity != "CallbackPreviewInput" || contract.CandidateEntity != "BoundedCallbackCandidate" || contract.EvidenceEntity != "CallbackPreviewEvidence" {
		t.Fatalf("callback preview contract identity = %#v", contract)
	}
	if len(contract.Fields) == 0 || len(contract.Activities) != 4 || len(contract.Bindings) != 3 {
		t.Fatalf("callback preview contract shape = fields:%d activities:%d bindings:%d", len(contract.Fields), len(contract.Activities), len(contract.Bindings))
	}
	for _, activity := range contract.Activities {
		if activity.InputEntity == "" || activity.OutputEntity == "" || !activity.UsedInputFact || !activity.GeneratedOutputFact {
			t.Fatalf("callback preview activity facts = %#v", activity)
		}
		if activity.OutputEntity == "OperationResult" {
			t.Fatalf("preview activity was bound to OperationResult: %#v", activity)
		}
	}
	fieldNames := make(map[string]bool, len(contract.Fields))
	for _, field := range contract.Fields {
		fieldNames[field.Entity+"."+field.Name] = true
		if field.ID == "" || field.TypeID == "" || field.Presence == "" || field.Cardinality == "" {
			t.Fatalf("incomplete callback preview field = %#v", field)
		}
	}
	for _, required := range []string{
		"CallbackPreviewInput.SourceDigest",
		"BoundedCallbackCandidate.CandidateDigest",
		"BoundedCallbackCandidate.PendingEffectCount",
		"CallbackCaptures.ObjectIdentities",
		"PendingCallbackEffects.Signatures",
		"CallbackPreviewEvidence.ApplyPermission",
	} {
		if !fieldNames[required] {
			t.Fatalf("missing typed preview field %q", required)
		}
	}
}

func TestCallbackPreviewContractRejectsMalformedSource(t *testing.T) {
	malformed := bytes.Replace(callbackPreviewContractSource, []byte("activity CloseCallbackPreview(PendingCallbackEffects)"), []byte("activity CloseCallbackPreview(MissingEntity)"), 1)
	if _, err := parseCallbackPreviewContract(malformed); err == nil {
		t.Fatal("malformed callback preview contract was accepted")
	}
}
