package cache

import (
	"fmt"
)

func validateEvidenceBindings(e EvidenceFreshness) error {
	bound := make(map[string]Digest, len(e.EvidenceRefs))
	for _, ref := range e.EvidenceRefs {
		bound[ref.Name] = ref.Digest
	}
	for _, required := range []struct {
		name   string
		digest Digest
	}{
		{"policy", e.PolicyDigest}, {"toolchain", e.ToolchainDigest},
	} {
		if bound[required.name] != required.digest {
			return fmt.Errorf("%w: evidence ref %q is not bound", ErrInvalidReceipt, required.name)
		}
	}
	return nil
}
