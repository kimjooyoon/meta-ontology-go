package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"path"
	"strings"
)

func validateProvenancePaths(paths []ProvenancePath) error {
	seenCommands := map[string]struct{}{}
	seenPaths := map[string]struct{}{}
	for _, path := range paths {
		if !validStableID(path.CommandID) || !validStableID(path.Requirement.PathID) {
			return failure(ReasonAmbiguousPath, "provenance identity is invalid")
		}
		if _, exists := seenCommands[path.CommandID]; exists {
			return failure(ReasonAmbiguousPath, "multiple provenance paths for command")
		}
		seenCommands[path.CommandID] = struct{}{}
		if _, exists := seenPaths[path.Requirement.PathID]; exists {
			return failure(ReasonAmbiguousPath, "duplicate provenance path ID")
		}
		seenPaths[path.Requirement.PathID] = struct{}{}
		if len(path.Requirement.RecordIDs) == 0 || len(path.Requirement.RecordIDs) != len(path.Requirement.ExpectedKinds) {
			return failure(ReasonAmbiguousPath, "provenance requirement is not finite")
		}
		if err := path.Path.Validate(); err != nil {
			return failure(ReasonAmbiguousPath, err.Error())
		}
		if err := validateIDs(append(append([]string{}, path.Requirement.RecordIDs...), path.Requirement.StartID, path.Requirement.EndID)); err != nil {
			return failure(ReasonAmbiguousPath, err.Error())
		}
	}
	return nil
}
func validateIDs(ids []string) error {
	for _, id := range ids {
		if !validStableID(id) {
			return fmt.Errorf("invalid stable ID %q", id)
		}
	}
	return nil
}
func validStableID(value string) bool {
	parsed, err := semantic.ParseIdentity(value)
	return err == nil && parsed.String() == value
}
func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
func validRepoPath(value string) bool {
	return value != "" && value != "." && !strings.HasPrefix(value, "/") && !strings.Contains(value, "\\") && path.Clean(value) == value && !strings.HasPrefix(value, "../") && value != ".." && !strings.ContainsRune(value, '\x00')
}
func validRepoDir(value string) bool {
	return value == "." || validRepoPath(value)
}
