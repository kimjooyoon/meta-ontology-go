package syntaxregistration

// ExecutionIdentity binds bytes, not paths or a version-only assertion.
// Environment configuration and publisher authenticity are separate obligations.
type ExecutionIdentity struct {
	GoVersion        string `json:"go_version"`
	GOOS             string `json:"goos"`
	GOARCH           string `json:"goarch"`
	ExecutableDigest string `json:"executable_sha256"`
	GoCommandDigest  string `json:"go_command_sha256"`
	CompilerDigest   string `json:"compiler_sha256"`
}

type ExecutionBinding struct {
	ActivityID string            `json:"activity_id"`
	InputID    string            `json:"input_id"`
	OutputID   string            `json:"output_id"`
	Identity   ExecutionIdentity `json:"identity"`
}

func validateExecutionIdentity(expected, observed ExecutionIdentity) error {
	for _, value := range []string{expected.GoVersion, expected.GOOS, expected.GOARCH,
		expected.ExecutableDigest, expected.GoCommandDigest, expected.CompilerDigest} {
		if value == "" {
			return failure("UNKNOWN", "bind-execution-identity", "REGISTRATION_EXECUTION_IDENTITY_MISSING",
				"DIRECT_MISSING", "observe-and-pin-execution-identity")
		}
	}
	for _, value := range []string{expected.ExecutableDigest, expected.GoCommandDigest, expected.CompilerDigest} {
		if !executionDigestValid(value) {
			return failure("REFUTED", "bind-execution-identity", "REGISTRATION_EXECUTION_IDENTITY_MALFORMED",
				"", "correct-explicit-execution-identity")
		}
	}
	if expected != observed {
		return failure("UNKNOWN", "bind-execution-identity", "REGISTRATION_EXECUTION_IDENTITY_STALE",
			"STALE", "observe-and-pin-execution-identity")
	}
	return nil
}

func executionDigestValid(value string) bool {
	if len(value) != 71 || value[:7] != "sha256:" {
		return false
	}
	for _, character := range value[7:] {
		if !('0' <= character && character <= '9') && !('a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}

func recheckExecutionIdentity(expected ExecutionIdentity) error {
	observed, err := ObserveExecutionIdentity()
	if err != nil {
		return err
	}
	return validateExecutionIdentity(expected, observed)
}
