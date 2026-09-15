package lanefrontier

import (
	"strings"
)

func baseInput() Input {
	return Input{
		SchemaVersion:     SchemaVersion,
		RegistryDigest:    strings.Repeat("a", 64),
		BaseSHA:           strings.Repeat("b", 40),
		LaneHeadSHA:       strings.Repeat("c", 40),
		LaneID:            "lane://billing",
		RegisteredBranch:  "agent/billing",
		OwnedPathPrefixes: []string{"internal/billing"},
		ChangedPaths:      []string{"internal/billing/order.go"},
	}
}
