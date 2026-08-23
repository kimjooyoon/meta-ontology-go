package languagegointeroperation

func executeGuardrailCase(definition Definition) CaseResult {
	if definition.Fixture == "unknown-payload" {
		evidence := rejectedEvidence("REGISTRY", "INTEROP_PAYLOAD_UNKNOWN")
		return finishCase(definition, evidence, definition.ExpectedStage == evidence.FailureStage)
	}
	source, found := guardrailFixture(definition.Fixture)
	if !found {
		evidence := rejectedEvidence("REGISTRY", "GUARDRAIL_FIXTURE_UNKNOWN")
		return finishCase(definition, evidence, false)
	}
	_, failure := inspectSource(source)
	if failure == nil {
		evidence := Evidence{ActualOutcome: "ACCEPT", InvalidAccepted: true,
			UnknownAccepted: definition.Fixture == "unknown-payload",
			ImportAccepted: definition.Fixture == "import-authority"}
		return finishCase(definition, evidence, false)
	}
	evidence := rejectedEvidence(failure.Stage, failure.Code)
	evidence.SourceDigest = digestBytes(source)
	satisfied := definition.ExpectedOutcome == "REJECT" && definition.ExpectedStage == failure.Stage
	return finishCase(definition, evidence, satisfied)
}
