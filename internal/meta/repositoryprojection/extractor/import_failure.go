package extractor

// Keep the owning source and bounded Go observation cause together. A missing
// observation still refuses extraction without refuting the source program.
func importResolutionFailure(root, logical, module, path string, cause error) error {
	diagnostics := projectedFailureDiagnostics(root, "logical="+logical, "import-resolution", preserveModuleFailure(cause))
	diagnostics = append(diagnostics, "module="+module, "import="+path)
	return Failure{
		Stage: "validate-ast-imports", Step: "resolve-import", Reason: "UNRESOLVED_IMPORT",
		UnknownClass: "DIRECT_MISSING", NextOperation: "restore-import-package",
		BlockedBy: []string{path}, Diagnostics: diagnostics,
	}
}
