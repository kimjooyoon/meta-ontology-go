package transformationeffect

// ValidateLedger exposes canonical ledger validation to read-only meta consumers.
func ValidateLedger(ledger Ledger) error {
	return validateLedger(ledger)
}
