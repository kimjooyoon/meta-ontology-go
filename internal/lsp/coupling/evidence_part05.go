package coupling

import (
	"fmt"
	"strings"
)

func validateRange(value Range) error {
	if value.Start.Line < 0 || value.Start.Character < 0 || value.End.Line < 0 || value.End.Character < 0 {
		return fmt.Errorf("range coordinates must be non-negative")
	}
	if value.End.Line < value.Start.Line || (value.End.Line == value.Start.Line && value.End.Character < value.Start.Character) {
		return fmt.Errorf("range end precedes start")
	}
	return nil
}
func exactText(value string) bool { return value != "" && strings.TrimSpace(value) == value }
