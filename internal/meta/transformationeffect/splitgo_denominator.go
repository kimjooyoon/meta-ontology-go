package transformationeffect

import (
	"errors"
	"fmt"
	"strings"
)

func validateSplitGoRequiredIDs(ids []string) error {
	if len(ids) == 0 {
		return errors.New("SplitGo indicator denominator is empty")
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return errors.New("SplitGo indicator denominator contains an empty id")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("SplitGo indicator denominator contains duplicate %q", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}
