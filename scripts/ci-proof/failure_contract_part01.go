package main

import (
	"fmt"
)

type failureCatalogEntry struct {
	Class          string
	Severity       string
	BlockingScope  string
	Parallelizable bool
	NextOperation  string
}
type failureCatalogRecord struct {
	Code  string
	Entry failureCatalogEntry
}

var failureCatalogRecords = []failureCatalogRecord{
	{Code: "CI-TEST-001", Entry: failureCatalogEntry{Class: "test", Severity: "error", BlockingScope: "local", Parallelizable: true, NextOperation: "REPAIR_SOURCE"}},
	{Code: "CI-SCOPE-001", Entry: failureCatalogEntry{Class: "scope", Severity: "error", BlockingScope: "global", Parallelizable: false, NextOperation: "RECOMPUTE_SCOPE"}},
	{Code: "CI-CAPS-001", Entry: failureCatalogEntry{Class: "caps", Severity: "error", BlockingScope: "global", Parallelizable: false, NextOperation: "REPAIR_SOURCE"}},
	{Code: "CI-CONTRACT-001", Entry: failureCatalogEntry{Class: "contract", Severity: "critical", BlockingScope: "global", Parallelizable: false, NextOperation: "RETAIN_UNRESOLVED"}},
	{Code: "CI-DEPENDENCY-001", Entry: failureCatalogEntry{Class: "dependency", Severity: "warning", BlockingScope: "local", Parallelizable: true, NextOperation: "CONTINUE_LOCAL"}},
	{Code: "CI-GATE-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "REOBSERVE_GATE"}},
	{Code: "CI-ARTIFACT-001", Entry: failureCatalogEntry{Class: "artifact", Severity: "error", BlockingScope: "global", Parallelizable: false, NextOperation: "REBUILD_ARTIFACT"}},
	{Code: "CI-FRESHNESS-001", Entry: failureCatalogEntry{Class: "freshness", Severity: "error", BlockingScope: "global", Parallelizable: false, NextOperation: "RERUN_CURRENT_HEAD"}},
	{Code: "CI-PROVENANCE-001", Entry: failureCatalogEntry{Class: "provenance", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "RECOMPUTE_PROVENANCE"}},
	{Code: "CI-PROMOTION-AUTH-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "REBUILD_PROMOTION_PROOF"}},
	{Code: "CI-PROMOTION-OBSERVATION-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "REOBSERVE_PROMOTION_TUPLE"}},
	{Code: "CI-ROOT-OF-TRUST-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "REVALIDATE_TRUST_ROOT"}},
	{Code: "CI-ROOT-OF-TRUST-BOOTSTRAP-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "REVALIDATE_TRUST_ROOT"}},
	{Code: "CI-UNCLASSIFIED-001", Entry: failureCatalogEntry{Class: "unclassified", Severity: "blocked", BlockingScope: "global", Parallelizable: false, NextOperation: "CLASSIFY_FROM_EVIDENCE"}},
}

func validFailureNextOperation(operation string) bool {
	switch operation {
	case "REPAIR_SOURCE", "RECOMPUTE_SCOPE", "RETAIN_UNRESOLVED", "CONTINUE_LOCAL", "REOBSERVE_GATE", "REBUILD_ARTIFACT", "RERUN_CURRENT_HEAD", "RECOMPUTE_PROVENANCE", "REBUILD_PROMOTION_PROOF", "REOBSERVE_PROMOTION_TUPLE", "REVALIDATE_TRUST_ROOT", "CLASSIFY_FROM_EVIDENCE":
		return true
	default:
		return false
	}
}

var failureCatalog = buildFailureCatalog()

func buildFailureCatalog() map[string]failureCatalogEntry {
	catalog := make(map[string]failureCatalogEntry, len(failureCatalogRecords))
	for _, record := range failureCatalogRecords {
		catalog[record.Code] = record.Entry
	}
	return catalog
}
func validateFailureCodes(codes []string, primary string) error {
	if len(codes) == 0 {
		return fmt.Errorf("failure code set is empty")
	}
	seen := make(map[string]bool, len(codes))
	for index, code := range codes {
		if _, ok := failureCatalog[code]; !ok || seen[code] || code == "" || (index > 0 && codes[index-1] >= code) {
			return fmt.Errorf("failure code set is unknown, duplicated, or not canonical")
		}
		seen[code] = true
	}
	if !seen[primary] {
		return fmt.Errorf("primary failure code is absent from the complete failure set")
	}
	return nil
}
