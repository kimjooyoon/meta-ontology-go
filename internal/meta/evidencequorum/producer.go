package evidencequorum

import (
	"encoding/json"
	"reflect"
)

type producerEntry struct {
	Package   string            `json:"package"`
	Namespace string            `json:"namespace"`
	Activity  string            `json:"activity"`
	Inputs    []producerBinding `json:"inputs"`
	Output    producerBinding   `json:"output"`
}

type producerBinding struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type producerEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}

type producerDiagnostic struct {
	Stage   string `json:"stage"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type producerEffects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type producerReceipt struct {
	Schema         string               `json:"schema"`
	Decision       string               `json:"decision"`
	Reason         string               `json:"reason"`
	Resolution     string               `json:"resolution"`
	Filename       string               `json:"filename"`
	SourceDigest   string               `json:"source_digest"`
	SemanticDigest string               `json:"semantic_digest,omitempty"`
	Entry          producerEntry        `json:"entry"`
	Events         []producerEvent      `json:"events"`
	Diagnostics    []producerDiagnostic `json:"diagnostics"`
	Effects        producerEffects      `json:"effects"`
	Digest         string               `json:"digest"`
}

func producerObservation(raw []byte, contract Contract, input Input) (Evidence, string, string, string) {
	receipt, err := decodeStrict[producerReceipt](raw)
	if err != nil || !verifyProducerReceipt(receipt) {
		return Evidence{}, "", StatusRefuted, "EVIDENCE_PRODUCER_RECEIPT_INVALID"
	}
	if receipt.Decision == DecisionUnknown {
		if receipt.Schema != "gooo/source-execution-receipt/v1" || receipt.Filename != input.SourcePath ||
			receipt.SourceDigest != SourceDigest(input.Source) || receipt.Entry.Activity != contract.SourceEntry {
			return Evidence{}, receipt.Digest, StatusRefuted, "EVIDENCE_PRODUCER_RECEIPT_MISMATCH"
		}
		return Evidence{}, receipt.Digest, StatusUnknown, "QUORUM_EVIDENCE_UNKNOWN"
	}
	if receipt.Decision != DecisionPass || receipt.Resolution != ResolutionExact ||
		receipt.Schema != "gooo/source-execution-receipt/v1" || receipt.Filename != input.SourcePath ||
		receipt.SourceDigest != SourceDigest(input.Source) || receipt.Reason != "SOURCE_ACTIVITY_EXECUTED" ||
		!validProducerEntry(receipt, contract) || !validProducerEvents(receipt, contract) ||
		len(receipt.Diagnostics) != 0 ||
		receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
		return Evidence{}, receipt.Digest, StatusRefuted, "EVIDENCE_PRODUCER_RECEIPT_MISMATCH"
	}
	return Evidence{ID: "producer-receipt", ClaimID: contract.Claim.ID, OriginGroup: "producer-run",
		Role: "producer", Producer: contract.Claim.Producer, Consumer: contract.Claim.Consumer,
		MetaOperation: contract.Claim.MetaOperation, ProofChoice: contract.Claim.ProofChoice,
		Value: "SUPPORTS", ConfidenceBPS: 10000, SourcePath: input.SourcePath,
		SourceDigest: SourceDigest(input.Source)}, receipt.Digest, StatusDischarged, ""
}

func validProducerEntry(receipt producerReceipt, contract Contract) bool {
	wantInputs := []producerBinding{{Name: "SourceProgram", ID: "evidence://source-program"},
		{Name: "Claim", ID: "evidence://claim"}}
	wantOutput := producerBinding{Name: "EvidenceReceipt", ID: "evidence://receipt"}
	return receipt.Entry.Package != "" && receipt.Entry.Namespace != "" && receipt.Entry.Activity == contract.SourceEntry &&
		reflect.DeepEqual(receipt.Entry.Inputs, wantInputs) && receipt.Entry.Output == wantOutput
}

func validProducerEvents(receipt producerReceipt, contract Contract) bool {
	wantKinds := []string{"SOURCE_PARSED", "SEMANTIC_LOWERED", "ACTIVITY_INVOKED", "ENTITY_PRODUCED"}
	wantSubjects := []string{receipt.SourceDigest, receipt.SemanticDigest, contract.SourceEntry, receipt.Entry.Output.ID}
	if len(receipt.Events) != len(wantKinds) || !validDigest(receipt.SemanticDigest) {
		return false
	}
	for index, event := range receipt.Events {
		if event.Sequence != index+1 || event.Kind != wantKinds[index] || event.Subject != wantSubjects[index] {
			return false
		}
	}
	return true
}

func verifyProducerReceipt(value producerReceipt) bool {
	digest := value.Digest
	value.Digest = ""
	return validDigest(digest) && digest == digestJSON(value)
}

func decodeProducerReceipt(raw []byte) (producerReceipt, error) {
	return decodeStrict[producerReceipt](raw)
}

func marshalProducerReceipt(value producerReceipt) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
