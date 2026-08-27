package evidencequorum

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestEvidenceQuorumUsesIndependentGroupsAndTransitions(t *testing.T) {
	report := Evaluate(fixtureInput())
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.RawEvidenceTotal != 12 || report.Summary.IndependentGroupsTotal != 11 ||
		report.Summary.DuplicateEvidenceTotal != 1 || report.Summary.ConflictCases != 1 {
		t.Fatalf("summary = %#v", report.Summary)
	}
	if report.Cases[1].IndependentGroups != 2 || report.Cases[1].DuplicateEvidence != 1 {
		t.Fatalf("replica case = %#v", report.Cases[1])
	}
	if report.Cases[2].ObservedReason != "QUORUM_CONFLICT" || report.Cases[2].ObservedResolution != ResolutionInvariant {
		t.Fatalf("conflict case = %#v", report.Cases[2])
	}
	if report.Cases[4].ObservedDecision != DecisionUnknown || report.Cases[4].Coordinate.Stage != "UNKNOWN" ||
		report.Cases[4].Coordinate.Step != "UNKNOWN" || report.Cases[4].Coordinate.Reason != "QUORUM_EVIDENCE_UNKNOWN" {
		t.Fatalf("unknown case = %#v", report.Cases[4])
	}
	for _, item := range report.Cases {
		if len(item.Claims) != 1 || len(item.Claims[0].Transitions) != 1 || item.Claims[0].Transitions[0].From != "OPEN" {
			t.Fatalf("claim transition = %#v", item.Claims)
		}
	}
}

func TestConfidenceDoesNotChangeQuorumDecision(t *testing.T) {
	input := fixtureInput()
	baseline := Evaluate(input)
	for _, receipts := range input.CaseReceipts {
		for index, raw := range receipts {
			receipt, err := DecodeReceipt(raw)
			if err != nil {
				t.Fatal(err)
			}
			for evidenceIndex := range receipt.Evidence {
				receipt.Evidence[evidenceIndex].ConfidenceBPS = 0
			}
			receipts[index] = marshalReceipt(SealReceipt(receipt))
		}
	}
	changed := Evaluate(input)
	for index := range baseline.Cases {
		if baseline.Cases[index].ObservedDecision != changed.Cases[index].ObservedDecision ||
			baseline.Cases[index].ObservedResolution != changed.Cases[index].ObservedResolution ||
			baseline.Cases[index].ObservedReason != changed.Cases[index].ObservedReason {
			t.Fatalf("confidence changed case %d: before=%#v after=%#v", index, baseline.Cases[index], changed.Cases[index])
		}
	}
	if changed.Summary.ConfidenceAggregated {
		t.Fatal("confidence aggregation was recorded as active")
	}
}

func fixtureInput() Input {
	contract := CanonicalContract()
	source := []byte("package evidencequorum\nnamespace evidencequorum\nactivity ProduceEvidence(SourceProgram, Claim) -> EvidenceReceipt\n")
	digest := SourceDigest(source)
	producer := producerReceipt{Schema: "gooo/source-execution-receipt/v1", Decision: DecisionPass,
		Reason: "SOURCE_ACTIVITY_EXECUTED", Resolution: ResolutionExact, Filename: contract.SourcePath,
		SourceDigest: digest, SemanticDigest: "sha256:" + strings.Repeat("c", 64),
		Entry: producerEntry{Package: "evidencequorum", Namespace: "evidencequorum", Activity: contract.SourceEntry,
			Inputs: []producerBinding{{Name: "SourceProgram", ID: "evidence://source-program"},
				{Name: "Claim", ID: "evidence://claim"}}, Output: producerBinding{Name: "EvidenceReceipt", ID: "evidence://receipt"}},
		Events: []producerEvent{{Sequence: 1, Kind: "SOURCE_PARSED", Subject: digest},
			{Sequence: 2, Kind: "SEMANTIC_LOWERED", Subject: "sha256:" + strings.Repeat("c", 64)},
			{Sequence: 3, Kind: "ACTIVITY_INVOKED", Subject: contract.SourceEntry},
			{Sequence: 4, Kind: "ENTITY_PRODUCED", Subject: "evidence://receipt"}},
		Diagnostics: []producerDiagnostic{}, Effects: producerEffects{}}
	producer.Digest = digestJSON(producer)
	unknownProducer := producer
	unknownProducer.Decision = DecisionUnknown
	unknownProducer.Digest = digestJSON(unknownProducer)
	newReceipt := func(id, role, group, value string, confidence int) []byte {
		claim := contract.Claim
		receipt := Receipt{Schema: ReceiptSchema, HeadSHA: strings.Repeat("a", 40),
			SourcePath: contract.SourcePath, SourceDigest: digest, Producer: claim.Producer,
			Consumer: claim.Consumer, MetaOperation: claim.MetaOperation, ProofChoice: claim.ProofChoice,
			Decision: DecisionPass, Resolution: ResolutionExact, Evidence: []Evidence{{
				ID: id, ClaimID: claim.ID, OriginGroup: group, Role: role, Producer: claim.Producer,
				Consumer: claim.Consumer, MetaOperation: claim.MetaOperation, ProofChoice: claim.ProofChoice,
				Value: value, ConfidenceBPS: confidence, SourcePath: contract.SourcePath, SourceDigest: digest,
			}}, RepositoryWrites: 0, MutationAuthority: false}
		return marshalReceipt(SealReceipt(receipt))
	}
	return Input{Contract: contract, HeadSHA: strings.Repeat("a", 40), SourcePath: contract.SourcePath,
		Source: source, CaseReceipts: [][][]byte{
			{newReceipt("producer-1", "producer", "producer-run", "SUPPORTS", 9100),
				newReceipt("consumer-1", "consumer", "consumer-check", "SUPPORTS", 8800),
				newReceipt("meta-1", "meta-operation", "quorum-meta", "SUPPORTS", 7600)},
			{newReceipt("producer-1", "producer", "producer-run", "SUPPORTS", 10000),
				newReceipt("producer-replica", "producer", "producer-run", "SUPPORTS", 10000),
				newReceipt("consumer-1", "consumer", "consumer-check", "SUPPORTS", 10000)},
			{newReceipt("producer-1", "producer", "producer-run", "SUPPORTS", 9900),
				newReceipt("consumer-1", "consumer", "consumer-check", "SUPPORTS", 9900),
				newReceipt("meta-1", "meta-operation", "quorum-meta", "SUPPORTS", 9900),
				newReceipt("contradictory-1", "consumer", "contradictory-check", "CONTRADICTS", 100)},
			{newReceipt("producer-1", "producer", "producer-run", "SUPPORTS", 10000),
				newReceipt("consumer-1", "consumer", "consumer-check", "SUPPORTS", 10000)},
			{}}, ProducerReceipt: marshalProducerReceipt(producer),
		UnknownProducerReceipt: marshalProducerReceipt(unknownProducer)}
}

func marshalReceipt(value Receipt) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("marshal receipt: %v", err))
	}
	return raw
}
