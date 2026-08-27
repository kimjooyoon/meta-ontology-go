package authorization

var caseIDs = []string{
	"exact", "envelope-unavailable", "report-unavailable", "report-decision-unknown",
	"policy-source-unavailable", "policy-generated-unavailable", "foundation-unavailable",
	"invocation-unavailable", "report-digest-mismatch", "subject-mismatch", "issuer-mismatch",
	"operation-mismatch", "scope-mismatch", "policy-mismatch", "default-allow",
	"run-id-mismatch", "run-attempt-mismatch", "nonce-mismatch", "effect-ceiling",
	"envelope-digest-mismatch",
}

func runCase(base Input, caseID string) Receipt {
	input := exactInput(base)
	switch caseID {
	case "exact":
	case "envelope-unavailable":
		input.EnvelopeAvailable = false
	case "report-unavailable":
		input.ReportAvailable = false
	case "report-decision-unknown":
		input.Report.Decision = "UNRECOGNIZED"
	case "policy-source-unavailable":
		input.Policy.SourceAvailable = false
	case "policy-generated-unavailable":
		input.Policy.GeneratedAvailable = false
	case "foundation-unavailable":
		input.Foundation.Available = false
	case "invocation-unavailable":
		input.Invocation.RunID = ""
	case "report-digest-mismatch":
		input.Envelope.SourceReportDigest = digestValue("mismatch")
		reseal(&input)
	case "subject-mismatch":
		input.Envelope.SubjectSHA = "0000000000000000000000000000000000000000"
		reseal(&input)
	case "issuer-mismatch":
		input.Envelope.Issuer = "github-actions:unknown"
		reseal(&input)
	case "operation-mismatch":
		input.Envelope.Operation = "gomacro.unbounded"
		reseal(&input)
	case "scope-mismatch":
		input.Envelope.Scope = "all"
		reseal(&input)
	case "policy-mismatch":
		input.Foundation.PolicySourceDigest = digestValue("mismatch")
	case "default-allow":
		input.Envelope.DefaultDecision = "ALLOW"
		reseal(&input)
	default:
		mutateLateCase(&input, caseID)
	}
	return Evaluate(input)
}

func mutateLateCase(input *Input, caseID string) {
	switch caseID {
	case "run-id-mismatch":
		input.Envelope.RunID = "different-run"
		reseal(input)
	case "run-attempt-mismatch":
		input.Envelope.RunAttempt++
		reseal(input)
	case "nonce-mismatch":
		input.Envelope.Nonce = digestValue("mismatch")
		input.Envelope.EnvelopeDigest = ""
		input.Envelope.EnvelopeDigest = digestValue(input.Envelope)
	case "effect-ceiling":
		input.Envelope.EffectCeiling.RepositoryWrites = 1
		reseal(input)
	case "envelope-digest-mismatch":
		input.Envelope.EnvelopeDigest = digestValue("mismatch")
	}
}

func reseal(input *Input) { input.Envelope = sealEnvelope(input.Envelope) }
