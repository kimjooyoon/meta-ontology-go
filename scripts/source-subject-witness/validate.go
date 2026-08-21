package main

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

func validateSource(report sourceReport, expectedSHA string) error {
	if report.Repository == "" || report.CommitSHA != expectedSHA || !validSHA(report.CommitSHA) {
		return fmt.Errorf("source identity is missing or not exact-head bound")
	}
	if report.Meta.Schema != "gooo/indicator-report/v3" || report.Meta.Policy.Schema != "gooo/source-policy/v1" {
		return fmt.Errorf("source metric schema is unsupported")
	}
	if !report.Meta.Policy.ExemptProjectRootTopology {
		return fmt.Errorf("project root topology exemption is required")
	}
	if err := validateFiles(report.Files); err != nil {
		return err
	}
	if err := validateDirectories(report.Directories); err != nil {
		return err
	}
	if err := validateDirectories(report.StorageDirectories); err != nil {
		return err
	}
	if len(report.Meta.Indicators) == 0 {
		return fmt.Errorf("source meta indicator ledger is empty")
	}
	for _, indicator := range report.Meta.Indicators {
		if indicator.Subject == "" || indicator.MetricID == "" || !indicator.Satisfied {
			return fmt.Errorf("source indicator is incomplete or unsatisfied")
		}
		if indicator.Decision != "PASS" && indicator.Decision != "NOT_APPLICABLE" {
			return fmt.Errorf("source indicator %q has decision %q", indicator.MetricID, indicator.Decision)
		}
	}
	return nil
}

func validateFiles(files []fileMetric) error {
	seen := make(map[string]bool)
	for _, file := range files {
		if !validPath(file.Path, false) || seen[file.Path] || file.Lines < 0 {
			return fmt.Errorf("file observation %q is invalid or duplicated", file.Path)
		}
		seen[file.Path] = true
		if file.Language != "go" && file.Language != "gooo" && file.Language != "other" {
			return fmt.Errorf("file %q has unknown language %q", file.Path, file.Language)
		}
		if file.Language == "other" && file.Lines != 0 {
			return fmt.Errorf("other file %q unexpectedly has language lines", file.Path)
		}
	}
	return nil
}

func validateDirectories(directories []directoryMetric) error {
	seen, roots := make(map[string]bool), 0
	for _, directory := range directories {
		if !validPath(directory.Path, true) || seen[directory.Path] {
			return fmt.Errorf("directory observation %q is invalid or duplicated", directory.Path)
		}
		seen[directory.Path] = true
		if directory.Path == "." { roots++; if directory.SubjectKind != "PROJECT_ROOT" { return fmt.Errorf("root kind is not PROJECT_ROOT") } } else if directory.SubjectKind != "DIRECTORY" { return fmt.Errorf("directory %q has kind %q", directory.Path, directory.SubjectKind) }
		for _, value := range []int{directory.DirectFolders, directory.DirectFiles, directory.RecursiveFolders, directory.RecursiveFiles, directory.GoFiles, directory.GoooFiles, directory.GoLines, directory.GoooLines} {
			if value < 0 { return fmt.Errorf("directory %q has a negative metric", directory.Path) }
		}
	}
	if roots != 1 { return fmt.Errorf("directory set has %d project roots", roots) }
	return nil
}

func validPath(path string, allowRoot bool) bool {
	if path == "." { return allowRoot }
	return path != "" && !filepath.IsAbs(path) && filepath.Clean(path) == path &&
		path != ".." && !strings.HasPrefix(path, "../") && filepath.ToSlash(path) == path
}

func validSHA(value string) bool {
	if len(value) != 40 { return false }
	_, err := hex.DecodeString(value)
	return err == nil
}
