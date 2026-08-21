package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type evaluationIssue struct {
	status Status
	code   ReasonCode
	detail string
}

func failIssue(code ReasonCode, detail string) *evaluationIssue {
	return &evaluationIssue{status: StatusFailClosed, code: code, detail: detail}
}
func unknownIssue(code ReasonCode, detail string) *evaluationIssue {
	return &evaluationIssue{status: StatusUnknown, code: code, detail: detail}
}
func required(fieldName string) *evaluationIssue {
	return unknownIssue(ReasonRequiredInputMissing, fieldName)
}
func normalizeID(id semantic.ID, fieldName string) (semantic.ID, *evaluationIssue) {
	parsed, err := semantic.ParseIdentity(id.String())
	if err != nil || parsed != id {
		return "", failIssue(ReasonMalformedBinding, fieldName)
	}
	return parsed, nil
}
func normalizeDigestValue(value, fieldName string) *evaluationIssue {
	if value == "" {
		return required(fieldName)
	}
	if !validDigest(value) {
		return failIssue(ReasonMalformedBinding, fieldName)
	}
	return nil
}
func normalizeConfig(config Config) *evaluationIssue {
	if config.Schema == "" || config.Baseline.Schema == "" {
		return required("config")
	}
	if config.Schema != ConfigSchemaV1 || config.Baseline.Schema != BaselineSchemaV1 {
		return failIssue(ReasonMalformedBinding, "config schema")
	}
	if !config.Baseline.FullSuiteRequired {
		return failIssue(ReasonMalformedBinding, "full-suite baseline is not enabled")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{config.RegistryDigest, "config registry digest"},
		{config.ToolchainDigest, "config toolchain digest"},
		{config.ProfileDigest, "config profile digest"},
		{config.SnapshotDigest, "config snapshot digest"},
		{config.ExpectedProviderDigest, "expected provider digest"},
		{config.ExpectedObserverDigest, "expected observer digest"},
		{config.Baseline.Digest, "baseline digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if stableDigest(baselineCanonical(config.Baseline)) != config.Baseline.Digest {
		return unknownIssue(ReasonStaleInput, "baseline digest")
	}
	return nil
}
