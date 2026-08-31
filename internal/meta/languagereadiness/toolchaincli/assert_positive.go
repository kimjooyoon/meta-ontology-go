package toolchaincli

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

const deferredProvenance = "gooo: provenance: deferred (no provenance records attached)\n"

func inspectPositive(operation string, observed cliruntime.Observation) outputEvidence {
	result := outputEvidence{stderrOK: observed.Stderr == ""}
	switch operation {
	case "VERSION_TEXT":
		result.stdoutOK = observed.Stdout == "gooo 0.2.0-dev (development)\n"
	case "VERSION_JSON":
		result.stdoutOK, result.structuredOutputs = validVersionJSON(observed.Stdout), 1
	case "CHECK_TEXT":
		result.stdoutOK = observed.Stdout == "ok: "+sourceFixture+"\n"
		result.languageOperations = 1
	case "CHECK_JSON":
		value, valid := decodeCommandJSON(observed.Stdout)
		result.stdoutOK = valid && validCommand(value, "check")
		result.structuredOutputs, result.languageOperations = 1, 1
	case "ROUNDTRIP_JSON":
		value, valid := decodeCommandJSON(observed.Stdout)
		result.stdoutOK = valid && validRoundtrip(value)
		result.structuredOutputs, result.languageOperations = 1, 1
	case "SEMANTIC_CHECK":
		result.stdoutOK = observed.Stdout == "ok: "+sourceFixture+"\n"
		result.stderrOK, result.languageOperations = observed.Stderr == deferredProvenance, 1
	}
	return result
}

func validCommand(value commandPayload, command string) bool {
	return value.SchemaVersion == "gooo/diagnostics/v1" && value.Command == command &&
		value.Status == "ok" && value.File == sourceFixture && len(value.Diagnostics) == 0
}

func validRoundtrip(value commandPayload) bool {
	return validCommand(value, "roundtrip") && value.OriginalHash != "" &&
		value.OriginalHash == value.RoundtripHash && value.Equivalent != nil && *value.Equivalent &&
		value.GetPut != nil && *value.GetPut && value.PutGet != nil && *value.PutGet
}
