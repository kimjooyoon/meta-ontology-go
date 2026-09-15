package packageexecution

import (
	"fmt"
	"strings"
)

func validateSources(values []SourceEvidence) error {
	for index, source := range values {
		if !strings.HasPrefix(source.Digest, "sha256:") || source.DeclarationCount < 0 {
			return fmt.Errorf("packageexecution: invalid source evidence")
		}
		if index > 0 && values[index-1].Filename >= source.Filename {
			return fmt.Errorf("packageexecution: sources are not strictly ordered")
		}
	}
	return nil
}
