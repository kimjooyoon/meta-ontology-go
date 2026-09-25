package metainvocation

import (
	"path"
	"strings"
)

func validateChangeSet(changeSet ChangeSet) (string, string) {
	if changeSet.Schema != InputSchema {
		return "INPUT_SCHEMA_MISMATCH", ""
	}
	if changeSet.CaseID == "" {
		return "CASE_ID_EMPTY", ""
	}
	if len(changeSet.Files) == 0 {
		return "CHANGE_SET_EMPTY", ""
	}
	seen := map[string]struct{}{}
	for _, file := range changeSet.Files {
		if file == "" {
			return "CHANGE_PATH_EMPTY", file
		}
		if path.IsAbs(file) || strings.HasPrefix(file, "\\") {
			return "CHANGE_PATH_ABSOLUTE", file
		}
		if strings.Contains(file, "\\") || path.Clean(file) != file || file == "." || strings.HasPrefix(file, "../") {
			return "CHANGE_PATH_NON_CANONICAL", file
		}
		if _, exists := seen[file]; exists {
			return "CHANGE_PATH_DUPLICATE", file
		}
		seen[file] = struct{}{}
	}
	return "", ""
}

func reasonWithFile(reason, file string) string {
	if file == "" {
		return reason
	}
	return reason + ":" + file
}
