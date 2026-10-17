package publictestreuse

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestPolicySHA256IdentityRetainsAlgorithmAndBytes(t *testing.T) {
	_, actual := valueexecution.RuntimeEvidenceIdentities()
	known := cache.HashBytes([]byte("evaluator executable")).String()
	cases := []struct {
		name     string
		identity string
		valid    bool
	}{
		{"runtime", actual, true},
		{"empty", "", false},
		{"missing-algorithm", known, false},
		{"wrong-algorithm", "sha512:" + known, false},
		{"uppercase", "sha256:" + strings.ToUpper(known), false},
		{"zero", "sha256:" + strings.Repeat("0", 64), false},
		{"short", "sha256:" + known[:63], false},
		{"trailing-space", "sha256:" + known + " ", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			digest, err := sha256IdentityDigest(test.identity)
			if test.valid {
				if err != nil || !cache.Digest(digest).Known() || "sha256:"+digest != test.identity {
					t.Fatalf("runtime identity bytes changed: %q, %v", digest, err)
				}
			} else if err == nil || digest != "" {
				t.Fatalf("unsupported identity acquired a digest: %q, %v", digest, err)
			}
		})
	}
}
