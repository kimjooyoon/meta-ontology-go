package artifactemit

func symbolicReaderIndicators(checks symbolicReaderChecks) []SymbolicValueContractIndicator {
	return []SymbolicValueContractIndicator{
		symbolicReaderIndicator("source.schema", "GUARDRAIL", "FOUNDATION", checks.Schema, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.subject", "GUARDRAIL", "FOUNDATION", checks.Subject, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.metric", "DRIVER", "FOUNDATION", checks.Metric, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.decision", "OUTCOME", "COHERENCE", checks.Decision, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.resolution", "GUARDRAIL", "COHERENCE", checks.Resolution, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.internal-digest", "GUARDRAIL", "FOUNDATION", checks.InternalDigest, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.upstream-digests", "GUARDRAIL", "FOUNDATION", checks.UpstreamDigests, "GOVERNOR"),
		symbolicReaderIndicator("source.unknown-branches", "GUARDRAIL", "REGRESSION", checks.UnknownBranches, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("source.unique-indicator-ids", "GUARDRAIL", "REGRESSION", checks.UniqueIndicatorID, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.user.present", "DRIVER", "FOUNDATION", checks.UserPresent, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.tool-author.present", "DRIVER", "FOUNDATION", checks.ToolPresent, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.governor.present", "DRIVER", "FOUNDATION", checks.GovernorPresent, "GOVERNOR"),
		symbolicReaderIndicator("reader.user.count-bound", "OUTCOME", "COHERENCE", checks.UserCountBound, "USER", "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.tool-author.count-bound", "OUTCOME", "COHERENCE", checks.ToolCountBound, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.governor.count-bound", "OUTCOME", "COHERENCE", checks.GovernorCountBound, "GOVERNOR"),
		symbolicReaderIndicator("reader.user-subset-tool-author", "DRIVER", "REGRESSION", checks.UserNested, "TOOL_AUTHOR", "GOVERNOR"),
		symbolicReaderIndicator("reader.tool-author-subset-governor", "DRIVER", "REGRESSION", checks.ToolNested, "GOVERNOR"),
		symbolicReaderIndicator("reader.resolutions-canonical", "GUARDRAIL", "COHERENCE", checks.ReaderResolutions, "USER", "TOOL_AUTHOR", "GOVERNOR"),
	}
}

func symbolicReaderIndicator(
	id, class, proof string,
	satisfied bool,
	audiences ...string,
) SymbolicValueContractIndicator {
	observed := 0
	if satisfied {
		observed = 1
	}
	return SymbolicValueContractIndicator{
		ID: id, Class: class, ProofChoice: proof, MetaOperation: "project-" + id,
		Observed: observed, Expected: 1, Satisfied: satisfied, Audiences: audiences,
	}
}
