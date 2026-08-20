package bidir

import (
	"errors"
	"slices"
)

func validateState(state BXStateEvidence) error {
	if state.Semantic == "" || state.Source == "" || state.Region == "" || state.Slot == "" || state.Bytes == "" || state.LStat == "" {
		return errors.New("semantic/source/region/slot/bytes/lstat digest is missing")
	}
	return nil
}
func containsString(values []string, want string) bool {
	return slices.Contains(values, want)
}
func sameIDs(left, right []ID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
