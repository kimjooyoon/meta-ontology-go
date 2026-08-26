package main

import "fmt"

func validateSource(report sourceReport, expectedSHA string) error {
	if report.Repository == "" || report.CommitSHA != expectedSHA || len(report.CommitSHA) != 40 {
		return fmt.Errorf("source identity is missing or not exact-head bound")
	}
	if report.Meta.Schema != "gooo/indicator-report/v3" || report.Meta.Policy.Schema != "gooo/source-policy/v1" {
		return fmt.Errorf("source metric schema is unsupported")
	}
	if !report.Meta.Policy.ExemptProjectRootTopology || !report.Meta.Policy.ExemptProjectRootREADME {
		return fmt.Errorf("project root topology and README exemptions are required")
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
		return fmt.Errorf("source meta indicator ledger is empty")
	}
	for _, indicator := range report.Meta.Indicators {
		validApplicability := indicator.Applicability == "APPLICABLE" || indicator.Applicability == "NOT_APPLICABLE"
		validDecision := indicator.Decision == "PASS" || indicator.Decision == "NOT_APPLICABLE"
		consistentApplicability := (indicator.Applicability == "NOT_APPLICABLE") == (indicator.Decision == "NOT_APPLICABLE")
		if indicator.Subject == "" || indicator.MetricID == "" || !indicator.Satisfied || !validApplicability || !validDecision || !consistentApplicability {
			return fmt.Errorf("source indicator %q is incomplete or unsatisfied", indicator.MetricID)
		}
	}
	return nil
}

func validateFiles(files []fileMetric) error {
	seen := make(map[string]bool)
	languages := map[string]bool{"go": true, "gooo": true, "other": true}
	for _, file := range files {
		invalidOther := file.Language == "other" && file.Lines != 0
		if !validPath(file.Path, false) || seen[file.Path] || file.Lines < 0 || !languages[file.Language] || invalidOther {
			return fmt.Errorf("file observation %q is invalid or duplicated", file.Path)
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
			return fmt.Errorf("directory observation %q is invalid or duplicated", directory.Path)
		}
		seen[directory.Path] = true
	}
	if roots != 1 {
		return fmt.Errorf("directory set has %d project roots", roots)
	}
	return nil
}
