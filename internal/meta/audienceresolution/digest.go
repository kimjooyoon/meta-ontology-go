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
	Source          SourceBinding    `json:"source"`
	Records         []EvidenceRecord `json:"records"`
	Counterexamples []Counterexample `json:"counterexamples"`
}

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	var normalized any
	if json.Unmarshal(raw, &normalized) != nil {
		return digestBytes(raw)
	}
	canonical, _ := json.Marshal(normalized)
	return digestBytes(canonical)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func factsDigest(ledger Ledger) string {
	return digestJSON(ledgerFacts{Schema: ledger.Schema, ID: ledger.ID, Subject: ledger.Subject,
		Source: ledger.Source, Records: ledger.Records, Counterexamples: ledger.Counterexamples})
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = ""
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func receiptDigest(receipt Receipt) string {
	// The seal covers the canonical receipt payload, excluding the field that
	// carries the seal itself. The independent consumer applies the same rule
	// to the raw receipt bytes.
	var value map[string]any
	raw, _ := json.Marshal(receipt)
	_ = json.Unmarshal(raw, &value)
	delete(value, "digest")
	return digestJSON(value)
}
