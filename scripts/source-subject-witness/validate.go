package main

func validateSource(report sourceReport, expectedSHA string) error {
	if report.Repository == "" || report.CommitSHA != expectedSHA || !validCommitSHA(report.CommitSHA) {
		return sourceValidationFailure("SOURCE_IDENTITY_MALFORMED", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	if report.Meta.Schema != "gooo/indicator-report/v3" || report.Meta.Policy.Schema != "gooo/source-policy/v1" {
		return sourceValidationFailure("SOURCE_SCHEMA_UNSUPPORTED", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	if !report.Meta.Policy.ExemptProjectRootTopology || !report.Meta.Policy.ExemptWorkflowDiscoveryRoot || !report.Meta.Policy.ExemptProjectRootREADME {
		return sourceValidationFailure("SOURCE_POLICY_INCOMPLETE", "MALFORMED_EVIDENCE", "restore-source-policy")
	}
	if report.Meta.Policy.MaxFileLines <= 0 || report.Meta.Policy.MaxFunctionLines <= 0 || report.Meta.Policy.MaxDirectDirectoryEntries < 0 {
		return sourceValidationFailure("SOURCE_POLICY_LIMITS_INVALID", "MALFORMED_EVIDENCE", "restore-source-policy")
	}
	for _, check := range []func() error{
		func() error { return validateFiles(report.Files) },
		func() error { return validateDirectories(report.Directories) },
		func() error { return validateDirectories(report.StorageDirectories) },
	} {
		if err := check(); err != nil {
			return err
		}
	}
	if len(report.Meta.Indicators) == 0 {
		return sourceValidationFailure("SOURCE_INDICATORS_MISSING", "DIRECT_MISSING", "restore-source-metrics")
	}
	seenIndicators := make(map[string]bool, len(report.Meta.Indicators))
	for _, indicator := range report.Meta.Indicators {
		if err := validateIndicatorShape(indicator); err != nil {
			return err
		}
		key := indicatorKey(indicator.SubjectKind, indicator.Subject, indicator.MetricID)
		if seenIndicators[key] {
			return sourceValidationFailure("SOURCE_INDICATOR_DUPLICATE", "MALFORMED_EVIDENCE", "restore-source-metrics")
		}
		seenIndicators[key] = true
		if err := validateIndicatorState(indicator); err != nil {
			return err
		}
	}
	return nil
}

func validateFiles(files []fileMetric) error {
	if len(files) == 0 {
		return sourceValidationFailure("SOURCE_FILES_MISSING", "DIRECT_MISSING", "restore-source-metrics")
	}
	seen := make(map[string]bool)
	languages := map[string]bool{"go": true, "gooo": true, "other": true}
	for _, file := range files {
		invalidOther := file.Language == "other" && file.Lines != 0
		if !validPath(file.Path, false) || seen[file.Path] || file.Lines < 0 || !languages[file.Language] || invalidOther {
			return sourceValidationFailure("SOURCE_FILE_OBSERVATION_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
		}
		seen[file.Path] = true
	}
	return nil
}

func validateDirectories(directories []directoryMetric) error {
	seen, roots := make(map[string]bool), 0
	for _, directory := range directories {
		kind := "DIRECTORY"
		if directory.Path == "." {
			roots++
			kind = "PROJECT_ROOT"
		}
		metricsValid := nonNegative(directory.DirectFolders, directory.DirectFiles,
			directory.RecursiveFolders, directory.RecursiveFiles, directory.GoFiles,
			directory.GoooFiles, directory.GoLines, directory.GoooLines)
		if !validPath(directory.Path, true) || seen[directory.Path] || directory.SubjectKind != kind || !metricsValid {
			return sourceValidationFailure("SOURCE_DIRECTORY_OBSERVATION_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
		}
		seen[directory.Path] = true
	}
	if roots != 1 {
		return sourceValidationFailure("SOURCE_DIRECTORY_ROOT_INVALID", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}
	return nil
}
