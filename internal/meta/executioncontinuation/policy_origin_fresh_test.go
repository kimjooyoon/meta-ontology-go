package executioncontinuation

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestUnavailableGrantPolicyFresh1625(t *testing.T) {
	f := fixtureConsumer1617(t)
	for _, c := range []struct {
		name string
		root fs.FS
	}{
		{"no-policy-root", nil}, {"missing-policy-file", fstest.MapFS{}},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateGrant(c.root, f.prepared.Pair.Grant, f.prepared.Metadata.Source)
			receipt := Diagnostic(f.prepared.Metadata.Source, err)
			if receipt.State != "DISPATCH_UNKNOWN" || receipt.Reason != "GRANT_POLICY_UNAVAILABLE" ||
				receipt.Unknown == nil || receipt.Unknown.Step != "policy_source" ||
				receipt.Unknown.Next != "restore_policy_source" {
				t.Fatal("missing policy confused with target run", receipt, err)
			}
		})
	}
}
