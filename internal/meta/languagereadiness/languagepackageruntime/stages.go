package languagepackageruntime

func stages(source Source, summary Summary) []Stage {
	definitions := []struct{ name, operation string }{
		{"registry", "decode-fixed-runtime-corpus"},
		{"manifest", "construct-package-runtime-manifest"},
		{"graph", "normalize-package-import-graph"},
		{"semantic", "lower-package-sources"},
		{"entry", "resolve-activity-contract"},
	}
	result := make([]Stage, len(definitions))
	input := digestValue(source)
	for index, definition := range definitions {
		output := digestValue(struct{ Previous, Operation string; Summary Summary }{input, definition.operation, summary})
		result[index] = Stage{Sequence: index + 1, Name: definition.name,
			MetaOperation: definition.operation, InputDigest: input, OutputDigest: output}
		input = output
	}
	return result
}
