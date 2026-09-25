package main

import (
	"fmt"
)

type failureCatalogEntry struct {
	Class           string
	Severity        string
	BlockingScope   string
	Parallelizable  bool
	HandoffRequired bool
	Owner           string
}
type failureCatalogRecord struct {
	Code  string
	Entry failureCatalogEntry
}

var failureCatalogRecords = []failureCatalogRecord{
	{Code: "CI-TEST-001", Entry: failureCatalogEntry{Class: "test", Severity: "error", BlockingScope: "local", Parallelizable: true, HandoffRequired: false, Owner: "registered-path-owner"}},
	{Code: "CI-SCOPE-001", Entry: failureCatalogEntry{Class: "scope", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false, Owner: "ci-policy"}},
	{Code: "CI-CAPS-001", Entry: failureCatalogEntry{Class: "caps", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false, Owner: "ci-policy"}},
	{Code: "CI-CONTRACT-001", Entry: failureCatalogEntry{Class: "contract", Severity: "critical", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-DEPENDENCY-001", Entry: failureCatalogEntry{Class: "dependency", Severity: "warning", BlockingScope: "local", Parallelizable: true, HandoffRequired: true, Owner: "registered-path-owner"}},
	{Code: "CI-GATE-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-ARTIFACT-001", Entry: failureCatalogEntry{Class: "artifact", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-FRESHNESS-001", Entry: failureCatalogEntry{Class: "freshness", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-PROVENANCE-001", Entry: failureCatalogEntry{Class: "provenance", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-OWNERSHIP-001", Entry: failureCatalogEntry{Class: "ownership", Severity: "blocked", BlockingScope: "local", Parallelizable: true, HandoffRequired: true, Owner: "branch-ownership"}},
	{Code: "CI-PROMOTION-AUTH-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-PROMOTION-OBSERVATION-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-ROOT-OF-TRUST-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-ROOT-OF-TRUST-BOOTSTRAP-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-UNCLASSIFIED-001", Entry: failureCatalogEntry{Class: "unclassified", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
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
