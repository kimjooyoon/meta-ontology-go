package toolchaincli

func stages(report Report) []Stage {
	registry := digestJSON(expectedRegistry())
	return []Stage{
		{Sequence: 1, Name: "registry", MetaOperation: "decode-fixed-cli-corpus",
			InputDigest: report.Source.RegistryDigest, OutputDigest: registry},
		{Sequence: 2, Name: "binary", MetaOperation: "bind-cli-binary-digest",
			InputDigest: registry, OutputDigest: report.Source.ExecutableDigest},
		{Sequence: 3, Name: "invoke", MetaOperation: "execute-cli-case-plan",
			InputDigest: report.Source.ExecutableDigest, OutputDigest: digestJSON(report.Cases)},
		{Sequence: 4, Name: "replay", MetaOperation: "compare-cli-observations",
			InputDigest: digestJSON(report.Cases), OutputDigest: digestJSON(report.Summary)},
		{Sequence: 5, Name: "effects", MetaOperation: "seal-cli-repository-effects",
			InputDigest: digestJSON(report.Summary), OutputDigest: digestJSON(report.Indicators)},
	}
}
