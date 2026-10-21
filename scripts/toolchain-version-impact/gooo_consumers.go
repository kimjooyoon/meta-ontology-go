package main

import (
	"regexp"
	"strings"
)

// These are version-bearing fields in Gooo computes metadata. A new field
// remains unresolved until its role and byte-level projection are understood.
var goooPatchVersion = regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+`)

type goooVersionSite struct {
	Start   int
	End     int
	Version string
	Role    string
	Valid   bool
}

func declaredGoooVersions(line string) []goooVersionSite {
	var sites []goooVersionSite
	offset := 0
	for segment := range strings.SplitSeq(line, ";") {
		trimmed := strings.TrimSpace(segment)
		matches := goooPatchVersion.FindAllStringIndex(segment, -1)
		role, valid := "", false
		switch {
		case strings.HasPrefix(trimmed, "compatibility-test-contract=go"):
			role = "gooo_compatibility_test_contract"
			valid = len(matches) == 1 &&
				strings.HasPrefix(trimmed, "compatibility-test-contract=go"+segment[matches[0][0]:matches[0][1]]+"|")
		case strings.HasPrefix(trimmed, "workflow-go-version="):
			role = "gooo_workflow_version"
			valid = len(matches) == 1 &&
				trimmed == "workflow-go-version="+segment[matches[0][0]:matches[0][1]]
		case strings.HasPrefix(trimmed, "badge-row=go-toolchain|"):
			role = "gooo_toolchain_badge"
			fields := strings.Split(trimmed, "|")
			if len(matches) == 3 && len(fields) >= 5 {
				version := segment[matches[0][0]:matches[0][1]]
				valid = true
				for _, field := range fields[2:5] {
					versions := goooPatchVersion.FindAllString(field, -1)
					valid = valid && len(versions) == 1 && versions[0] == version
				}
			}
		}
		if role != "" {
			for _, match := range matches {
				sites = append(sites, goooVersionSite{
					Start: offset + match[0], End: offset + match[1],
					Version: segment[match[0]:match[1]], Role: role, Valid: valid,
				})
			}
		}
		offset += len(segment) + 1
	}
	return sites
}
