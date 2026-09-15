package lanefrontier

import (
	"testing"
)

func FuzzEvaluateNeverPanics(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"schema":"gooo/lane-frontier/v1","changed_paths":[]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		input, err := DecodeInput(data)
		if err == nil {
			_ = Evaluate(input)
		}
	})
}
func permutationCase() Case {
	return Case{Name: "permutation", Input: Input{Schema: SchemaV1, RegistryDigest: "registry", BaseSHA: "base", LaneHeadSHA: "head", LaneStableID: "lane", RegisteredBranch: "agent/fuzz-conformance", OwnedPathPrefixes: []string{"pkg/a", "pkg/z"}, ChangedPaths: []string{"pkg/a/file.go", "pkg/z/file.go"}}, Expected: Result{Decision: Eligible, Reason: Clean}}
}
