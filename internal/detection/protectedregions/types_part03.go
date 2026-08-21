package protectedregions

import (
	"fmt"
	"strings"
)

// Valid reports whether both sources are structurally sound and the change is
// confined to generated region bodies.
func (r LocalityReport) Valid() bool {
	return r.Before.Valid() && r.After.Valid() && len(r.Violations) == 0
}

// Err turns a locality report into a fail-fast error.
func (r LocalityReport) Err() error {
	if err := r.Before.Err(); err != nil {
		return err
	}
	if err := r.After.Err(); err != nil {
		return err
	}
	if len(r.Violations) == 0 {
		return nil
	}
	lines := make([]string, len(r.Violations))
	for index, violation := range r.Violations {
		lines[index] = violation.Error()
	}
	return fmt.Errorf("protected-region locality validation failed:\n%s", strings.Join(lines, "\n"))
}
