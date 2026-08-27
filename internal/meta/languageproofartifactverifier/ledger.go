package languageproofartifactverifier

import "reflect"

const LedgerSchema = "gooo/proof-evidence-ledger/v2"

func validatePriorLedger(ledger ClaimLedger, claims []ClaimStatement) error {
	if err := validateOpenLedger(ledger); err != nil || len(claims) != ClaimTemplateTotal {
		return errLedger("prior ledger identity")
	}
	previous := ""
	for index, statement := range claims {
		entry := ledger.Entries[index]
		if entry.ClaimID != statement.ID || entry.Proposition != statement.Proposition || entry.TargetDigest != statement.TargetDigest || !reflect.DeepEqual(entry.Dependencies, statement.Dependencies) ||
			entry.Status != "OPEN" || entry.Resolution != "LOWER_RESOLUTION" || entry.Producer != ProducerID || entry.Consumer != ConsumerID ||
			entry.ProofChoice != statement.ProofChoice || entry.MetaOperation != statement.MetaOperation || entry.Coordinate != statement.Coordinate ||
			entry.Reason != "AWAITING_INDEPENDENT_RECHECK" || !reflect.DeepEqual(entry.EvidenceDigest, statement.EvidenceDigest) ||
			entry.Provenance != "producer-carried-prior-ledger" || entry.PreviousDigest != previous || entry.Digest != ledgerEntryDigest(entry) {
			return errLedger("prior ledger entry")
		}
		previous = entry.Digest
	}
	return nil
}

func validateOpenLedger(ledger ClaimLedger) error {
	if ledger.Schema != LedgerSchema || ledger.Version != 2 || len(ledger.Entries) != ClaimTemplateTotal || ledger.Digest != claimLedgerDigest(ledger) {
		return errLedger("open ledger identity")
	}
	previous := ""
	for _, entry := range ledger.Entries {
		if entry.Status != "OPEN" || entry.Resolution != "LOWER_RESOLUTION" || entry.PreviousDigest != previous || entry.Digest != ledgerEntryDigest(entry) {
			return errLedger("open ledger chain")
		}
		previous = entry.Digest
	}
	return nil
}

func validateFinalLedger(ledger, prior ClaimLedger) error {
	if ledger.Schema != LedgerSchema || ledger.Version != 2 || len(ledger.Entries) != ClaimTemplateTotal*2 || ledger.Digest != claimLedgerDigest(ledger) {
		return errLedger("final ledger identity")
	}
	if err := validateOpenLedger(prior); err != nil {
		return err
	}
	for index := range prior.Entries {
		if !reflect.DeepEqual(ledger.Entries[index], prior.Entries[index]) {
			return errLedger("final ledger discarded prior entries")
		}
	}
	previous := lastDigest(prior.Entries)
	for index, entry := range ledger.Entries[ClaimTemplateTotal:] {
		priorEntry := prior.Entries[index]
		if entry.ClaimID != priorEntry.ClaimID || entry.Proposition != priorEntry.Proposition || entry.TargetDigest != priorEntry.TargetDigest || !reflect.DeepEqual(entry.Dependencies, priorEntry.Dependencies) ||
			entry.Status != "DISCHARGED" || entry.Resolution != "EXACT" || !reflect.DeepEqual(entry.EvidenceDigest, priorEntry.EvidenceDigest) ||
			entry.PreviousDigest != previous || entry.Digest != ledgerEntryDigest(entry) {
			return errLedger("final ledger discharge chain")
		}
		previous = entry.Digest
	}
	return nil
}

func appendLedgerEntry(ledger ClaimLedger, claim ClaimResult) ClaimLedger {
	entry := LedgerEntry{ClaimID: claim.ID, Proposition: claim.Proposition, TargetDigest: claim.TargetDigest, Dependencies: append([]string(nil), claim.Dependencies...), Status: "DISCHARGED", Resolution: "EXACT",
		Producer: ProducerID, Consumer: ConsumerID, ProofChoice: claim.ProofChoice, MetaOperation: claim.MetaOperation, Coordinate: claim.Coordinate,
		Reason: "INDEPENDENT_SOURCE_OPERATION_RECIPE_RECHECKED", EvidenceDigest: append([]string(nil), claim.EvidenceDigests...), Provenance: "consumer-canonical-recipe-v2", PreviousDigest: lastDigest(ledger.Entries)}
	entry.Digest = ledgerEntryDigest(entry)
	ledger.Entries = append(ledger.Entries, entry)
	ledger.Digest = claimLedgerDigest(ledger)
	return ledger
}

func dischargeLedger(prior ClaimLedger, cases []CaseResult) ClaimLedger {
	ledger := prior
	for _, item := range cases {
		if item.ID != "valid-proof-carrying-artifact" || item.ObservedDecision != "PASS" {
			continue
		}
		for _, claim := range item.Claims {
			if claim.Status == "DISCHARGED" {
				ledger = appendLedgerEntry(ledger, claim)
			}
		}
		break
	}
	return ledger
}

func claimLedgerDigest(ledger ClaimLedger) string {
	ledger.Digest = ""
	return digestValue(ledger)
}

func ledgerEntryDigest(entry LedgerEntry) string {
	entry.Digest = ""
	return digestValue(entry)
}

func lastDigest(entries []LedgerEntry) string {
	if len(entries) == 0 {
		return ""
	}
	return entries[len(entries)-1].Digest
}

func errLedger(reason string) error { return &ledgerError{reason: reason} }

type ledgerError struct{ reason string }

func (e *ledgerError) Error() string { return "proof evidence ledger mismatch: " + e.reason }
