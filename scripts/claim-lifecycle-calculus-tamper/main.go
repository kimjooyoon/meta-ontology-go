package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// This fixture is deliberately not a judge. It changes one ledger event and
// reseals its cause and append-only chain, proving that source reconstruction,
// rather than a broken digest alone, is the rejection boundary.
type Evidence struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	ClaimID      string `json:"claim_id"`
	SourceCaseID string `json:"source_case_id"`
	Digest       string `json:"digest"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type CauseReceipt struct {
	Sequence           int        `json:"sequence"`
	ClaimID            string     `json:"claim_id"`
	Kind               string     `json:"kind"`
	EvidenceIDs        []string   `json:"evidence_ids,omitempty"`
	DependencyClaimIDs []string   `json:"dependency_claim_ids,omitempty"`
	Coordinate         Coordinate `json:"coordinate"`
	Reason             string     `json:"reason"`
	Provenance         string     `json:"provenance"`
	Digest             string     `json:"digest"`
}

type Transition struct {
	Sequence           int    `json:"sequence"`
	ClaimID            string `json:"claim_id"`
	DeclarationDigest  string `json:"declaration_digest"`
	Event              string `json:"event"`
	Before             string `json:"before"`
	After              string `json:"after"`
	EvidenceDigest     string `json:"evidence_digest,omitempty"`
	CauseReceiptDigest string `json:"cause_receipt_digest"`
	PreviousDigest     string `json:"previous_digest,omitempty"`
	Digest             string `json:"digest"`
}

// RawMessage keeps untouched receipt sections opaque while preserving the
// producer's field order for the outer digest.
type Envelope struct {
	Schema                string          `json:"schema"`
	Scope                 string          `json:"scope"`
	HeadSHA               string          `json:"head_sha"`
	GoVersion             string          `json:"go_version"`
	SourcePath            string          `json:"source_path"`
	RawSourceDigest       string          `json:"raw_source_digest"`
	Producer              string          `json:"producer"`
	Consumer              string          `json:"consumer"`
	MetaOperation         string          `json:"meta_operation"`
	SourceRelation        json.RawMessage `json:"source_relation"`
	Claims                json.RawMessage `json:"claims"`
	Evidence              []Evidence      `json:"evidence"`
	Transitions           []Transition    `json:"transitions"`
	CauseReceipts         []CauseReceipt  `json:"cause_receipts"`
	Cases                 json.RawMessage `json:"cases"`
	Metrics               json.RawMessage `json:"metrics"`
	Summary               json.RawMessage `json:"summary"`
	Effects               json.RawMessage `json:"effects"`
	ConformanceDecision   json.RawMessage `json:"conformance_decision"`
	SubjectCounts         json.RawMessage `json:"subject_counts"`
	SubjectResolution     json.RawMessage `json:"subject_resolution"`
	SemanticReceiptDigest string          `json:"semantic_receipt_digest"`
	ReceiptDigest         string          `json:"receipt_digest"`
}

func main() {
	input := flag.String("input", "", "baseline receipt")
	output := flag.String("output", "", "coherently resealed tamper output")
	flag.Parse()
	if *input == "" || *output == "" {
		fail("-input and -output are required")
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		fail(err.Error())
	}
	var envelope Envelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		fail(err.Error())
	}
	var supporting Evidence
	for _, evidence := range envelope.Evidence {
		if evidence.Kind == "SUPPORTING" {
			supporting = evidence
			break
		}
	}
	if supporting.ID == "" || len(envelope.Transitions) < 9 || len(envelope.CauseReceipts) < 9 {
		fail("baseline receipt has no supporting evidence or contradiction transition")
	}
	const contradictionIndex = 8
	cause := &envelope.CauseReceipts[contradictionIndex]
	cause.Kind = "SUPPORTING_EVIDENCE"
	cause.EvidenceIDs = []string{supporting.ID}
	cause.Coordinate.Reason = "SUPPORTING_EVIDENCE"
	cause.Reason = "SUPPORTING_EVIDENCE"
	cause.Digest = digestWithoutCause(*cause)
	transition := &envelope.Transitions[contradictionIndex]
	transition.Event = "EVIDENCE_ACCEPTED"
	transition.After = "DISCHARGED"
	transition.EvidenceDigest = supporting.Digest
	transition.CauseReceiptDigest = cause.Digest
	for index := contradictionIndex; index < len(envelope.Transitions); index++ {
		if index == 0 {
			envelope.Transitions[index].PreviousDigest = ""
		} else {
			envelope.Transitions[index].PreviousDigest = envelope.Transitions[index-1].Digest
		}
		envelope.Transitions[index].Digest = digestWithoutTransition(envelope.Transitions[index])
	}
	envelope.ReceiptDigest = digestWithoutReceipt(envelope)
	encoded, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fail(err.Error())
	}
}

func digestWithoutCause(value CauseReceipt) string {
	value.Digest = ""
	return digestJSON(value)
}

func digestWithoutTransition(value Transition) string {
	value.Digest = ""
	return digestJSON(value)
}

func digestWithoutReceipt(value Envelope) string {
	value.ReceiptDigest = ""
	return digestJSON(value)
}

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
