package languageproofartifact

const LedgerSchema = "gooo/proof-evidence-ledger/v1"

func openLedger(evidence []Evidence) ClaimLedger {
	ledger := ClaimLedger{Schema: LedgerSchema, Version: 1, Entries: []LedgerEntry{}}
	for _, item := range evidence {
		entry := LedgerEntry{ClaimID: item.ClaimID, Status: "OPEN", Resolution: "LOWER_RESOLUTION",
			Producer: ProducerID, Consumer: ConsumerID, ProofChoice: item.ProofChoice,
			MetaOperation: item.MetaOperation, Coordinate: item.Coordinate,
			Reason: "AWAITING_INDEPENDENT_RECHECK", EvidenceDigest: []string{item.EvidenceDigest},
			Provenance: "producer-carried-prior-ledger", PreviousDigest: lastDigest(ledger.Entries)}
		entry.Digest = ledgerEntryDigest(entry)
		ledger.Entries = append(ledger.Entries, entry)
	}
	ledger.Digest = claimLedgerDigest(ledger)
	return ledger
}

func ledgerEntryDigest(entry LedgerEntry) string {
	entry.Digest = ""
	return digestValue(entry)
}

func claimLedgerDigest(ledger ClaimLedger) string {
	ledger.Digest = ""
	return digestValue(ledger)
}

func lastDigest(entries []LedgerEntry) string {
	if len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Digest
}
