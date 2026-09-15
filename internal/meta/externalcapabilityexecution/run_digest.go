package externalcapabilityexecution

func normalizedRunDigest(run CapabilityRun) string {
	normalized := struct {
		Status                string `json:"status"`
		Arithmetic            string `json:"arithmetic"`
		Function              string `json:"function"`
		EvaluatorExitCode     int    `json:"evaluator_exit_code"`
		EvaluatorOutputBytes  int    `json:"evaluator_output_bytes"`
		EvaluatorOutputSHA256 string `json:"evaluator_output_sha256"`
		MacroExitCode         int    `json:"macro_exit_code"`
		MacroGeneratedSHA256  string `json:"macro_generated_sha256"`
		MacroExpectedSHA256   string `json:"macro_expected_sha256"`
	}{
		run.Status, run.Arithmetic, run.Function, run.EvaluatorExitCode,
		run.EvaluatorOutputBytes, run.EvaluatorOutputSHA256, run.MacroExitCode,
		run.MacroGeneratedSHA256, run.MacroExpectedSHA256,
	}
	return digestValue(normalized)
}

func status(exact bool) string {
	if exact {
		return StatusSatisfied
	}
	return StatusUnsatisfied
}
