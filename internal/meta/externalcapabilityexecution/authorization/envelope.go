package authorization

func Issue(subject, runID string, runAttempt int, reportDigest string,
	policy PolicyEvidence) Envelope {
	envelope := Envelope{
		Schema: EnvelopeSchema, SubjectSHA: subject, Issuer: ExpectedIssuer,
		Operation: ExpectedOperation, Scope: ExpectedScope,
		PolicySourceDigest: policy.SourceDigest,
		PolicyGeneratedDigest: policy.GeneratedDigest,
		SourceReportDigest: reportDigest, DefaultDecision: ExpectedDefaultDecision,
		RunID: runID, RunAttempt: runAttempt, EffectCeiling: EffectCeiling{},
	}
	return sealEnvelope(envelope)
}

func expectedNonce(envelope Envelope) string {
	envelope.Nonce = ""
	envelope.EnvelopeDigest = ""
	return digestValue(envelope)
}

func sealEnvelope(envelope Envelope) Envelope {
	envelope.Nonce = expectedNonce(envelope)
	envelope.EnvelopeDigest = ""
	envelope.EnvelopeDigest = digestValue(envelope)
	return envelope
}

func envelopeSealExact(envelope Envelope) bool {
	wantNonce := expectedNonce(envelope)
	digest := envelope.EnvelopeDigest
	envelope.EnvelopeDigest = ""
	return envelope.Nonce == wantNonce && digest == digestValue(envelope)
}
