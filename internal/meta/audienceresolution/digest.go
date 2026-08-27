package audienceresolution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type ledgerFacts struct {
	Schema          string           `json:"schema"`
	ID              string           `json:"id"`
	Subject         string           `json:"subject"`
	Decision        string           `json:"decision"`
	Resolution      string           `json:"resolution"`
	Reason          string           `json:"reason"`
	Source          SourceBinding    `json:"source"`
	Records         []EvidenceRecord `json:"records"`
	Counterexamples []Counterexample `json:"counterexamples"`
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func factsDigest(ledger Ledger) string {
	return digestJSON(ledgerFacts{Schema: ledger.Schema, ID: ledger.ID, Subject: ledger.Subject,
		Decision: ledger.Decision, Resolution: ledger.Resolution, Reason: ledger.Reason,
		Source: ledger.Source, Records: ledger.Records, Counterexamples: ledger.Counterexamples})
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = digestJSON(receipt)
	return receipt
}
