package verify

import (
	"strings"
	"testing"
)

func TestRefusalRegistryScopeFresh1739(t *testing.T) {
	for _, row := range []struct {
		name, operation               string
		declared, absent, unsupported bool
	}{
		{"both-known", "bind-exact-source-metrics", true, false, false},
		{"kernel-only", "bind-exact-source-metrics", false, true, false},
		{"source-only", "gooo://new/source-operation", true, false, true},
		{"neither-known", "gooo://new/source-operation", false, true, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			absent, err := refusalSourceAbsent(row.operation, map[string]bool{row.operation: row.declared})
			if absent != row.absent || (err != nil) != row.unsupported ||
				(err != nil && !strings.HasPrefix(err.Error(), "COMPATIBILITY_KERNEL_UNSUPPORTED:")) {
				t.Fatalf("source absence and kernel support conflated: absent=%v err=%v", absent, err)
			}
		})
	}
}
