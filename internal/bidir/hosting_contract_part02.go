package bidir

import (
	"fmt"
	"sort"
	"strings"
)

// Validate checks phase identity and evidence uniqueness.
func (c HostingContract) Validate() error {
	if c.Phase != HostPhaseGoHosted && c.Phase != HostPhaseGoooHosted {
		return fmt.Errorf("unknown host phase %d", c.Phase)
	}
	if strings.TrimSpace(c.HostLanguage) == "" {
		return fmt.Errorf("host language is required")
	}
	if strings.TrimSpace(c.AuthoritativeView) == "" {
		return fmt.Errorf("authoritative view is required")
	}
	seen := make(map[string]struct{}, len(c.Evidence))
	for _, evidence := range c.Evidence {
		check := strings.TrimSpace(evidence.Check)
		if check == "" {
			return fmt.Errorf("evidence check is required")
		}
		if evidence.State < EvidencePlanned || evidence.State > EvidenceVerified {
			return fmt.Errorf("evidence %q has unknown state %d", check, evidence.State)
		}
		if _, exists := seen[check]; exists {
			return fmt.Errorf("duplicate evidence check %q", check)
		}
		seen[check] = struct{}{}
	}
	return nil
}

// Verified reports whether every declared check has verified evidence.
func (c HostingContract) Verified() bool {
	if c.Validate() != nil || len(c.Evidence) == 0 {
		return false
	}
	for _, evidence := range c.Evidence {
		if evidence.State != EvidenceVerified {
			return false
		}
	}
	return true
}

// UnverifiedChecks returns planned or observed checks in stable order.
func (c HostingContract) UnverifiedChecks() []string {
	var checks []string
	for _, evidence := range c.Evidence {
		if evidence.State != EvidenceVerified {
			checks = append(checks, strings.TrimSpace(evidence.Check))
		}
	}
	sort.Strings(checks)
	return checks
}

// HostingComparison describes the evidence gap between two phases.
type HostingComparison struct {
	From          HostPhase
	To            HostPhase
	HostChanged   bool
	AddedChecks   []string
	NewlyVerified []string
	Remaining     []string
}
