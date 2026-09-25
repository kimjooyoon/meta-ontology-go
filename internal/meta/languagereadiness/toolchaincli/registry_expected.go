package toolchaincli

func expectedRegistry() Registry {
	return Registry{Schema: RegistrySchema, Version: RegistryVersion, Cases: []Definition{
		{ID: "positive-version-text", Kind: "POSITIVE", Operation: "VERSION_TEXT", ExpectedExit: 0, ProofChoice: "FOUNDATION", MetaOperation: "observe-cli-identity"},
		{ID: "positive-version-json", Kind: "POSITIVE", Operation: "VERSION_JSON", ExpectedExit: 0, ProofChoice: "FOUNDATION", MetaOperation: "decode-cli-identity"},
		{ID: "positive-check-text", Kind: "POSITIVE", Operation: "CHECK_TEXT", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "execute-syntax-check"},
		{ID: "positive-check-json", Kind: "POSITIVE", Operation: "CHECK_JSON", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "decode-syntax-check"},
		{ID: "positive-roundtrip-json", Kind: "POSITIVE", Operation: "ROUNDTRIP_JSON", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "execute-roundtrip-check"},
		{ID: "positive-semantic-check", Kind: "POSITIVE", Operation: "SEMANTIC_CHECK", ExpectedExit: 0, ProofChoice: "COHERENCE", MetaOperation: "execute-semantic-check"},
		{ID: "guardrail-root-usage", Kind: "GUARDRAIL", Operation: "ROOT_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-missing-command"},
		{ID: "guardrail-unknown-command", Kind: "GUARDRAIL", Operation: "UNKNOWN_COMMAND", ExpectedExit: 1, ProofChoice: "REGRESSION", MetaOperation: "reject-unknown-command"},
		{ID: "guardrail-version-usage", Kind: "GUARDRAIL", Operation: "VERSION_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-version-argument"},
		{ID: "guardrail-check-usage", Kind: "GUARDRAIL", Operation: "CHECK_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-missing-check-input"},
		{ID: "guardrail-roundtrip-usage", Kind: "GUARDRAIL", Operation: "ROUNDTRIP_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-missing-roundtrip-input"},
		{ID: "guardrail-lsp-usage", Kind: "GUARDRAIL", Operation: "LSP_USAGE", ExpectedExit: 2, ProofChoice: "REGRESSION", MetaOperation: "reject-lsp-argument"},
	}}
}
