package transformationeffect

// ValidateLedger exposes canonical ledger validation to read-only meta consumers.
func ValidateLedger(ledger Ledger) error {
	if err := validateLedger(ledger); err != nil {
		return err
	}
	return validateSplitGoEffects(ledger.Effects)
}
