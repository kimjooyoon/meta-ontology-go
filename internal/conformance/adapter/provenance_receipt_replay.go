package adapter

// ReceiptReplayGuard checks append-only history for event identity reuse.
// It is caller-fed history; it does not infer external CI or store state.
type ReceiptReplayGuard struct {
	events map[string]string
}

// NewReceiptReplayGuard creates an empty append-only receipt history.
func NewReceiptReplayGuard() *ReceiptReplayGuard {
	return &ReceiptReplayGuard{events: make(map[string]string)}
}

// Accept validates and records one receipt, rejecting event_ref replay.
func (g *ReceiptReplayGuard) Accept(receipt ProvenanceReceipt) error {
	if g == nil {
		return oracleError(OracleNW003, "receipt replay guard is missing")
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	return g.acceptValidated(receipt)
}

// AcceptObserved requires observer-owned no-write proof before recording a receipt.
func (g *ReceiptReplayGuard) AcceptObserved(
	request Request, receipt ProvenanceReceipt, observation *NoWriteObservation,
) error {
	if err := receipt.ValidateObservedNoWrite(request, observation); err != nil {
		return err
	}
	return g.acceptValidated(receipt)
}

func (g *ReceiptReplayGuard) acceptValidated(receipt ProvenanceReceipt) error {
	if g == nil {
		return oracleError(OracleNW003, "receipt replay guard is missing")
	}
	digest, err := receipt.Digest()
	if err != nil {
		return err
	}
	if g.events == nil {
		g.events = make(map[string]string)
	}
	if previous, exists := g.events[receipt.EventRef]; exists {
		if previous == digest {
			return oracleError(OracleNW003, "receipt event_ref was replayed")
		}
		return oracleError(OracleNW003, "receipt event_ref has a conflicting digest")
	}
	g.events[receipt.EventRef] = digest
	return nil
}

// ValidateAppendOnly validates a receipt against caller-supplied prior history.
func (r ProvenanceReceipt) ValidateAppendOnly(previous []ProvenanceReceipt) error {
	guard := NewReceiptReplayGuard()
	for _, prior := range previous {
		if err := guard.Accept(prior); err != nil {
			return err
		}
	}
	return guard.Accept(r)
}
