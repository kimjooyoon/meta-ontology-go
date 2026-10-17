package publictestreuse

import (
	"bytes"
	"os"
	"testing"
)

func TestPolicyRequiresObservedPackageContract(t *testing.T) {
	filename := "../../../examples/self-improvement-discovery/project.gooo"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Load(filename, source)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Schema != PolicySchema || len(policy.Bindings) != 12 || policy.Bindings[10] != "execution-package-inputs" {
		t.Fatal("current policy omitted observed execution package inputs")
	}
	for _, marker := range []string{"test-reuse-contract=v1", "test-reuse-contract="} {
		changed := bytes.Replace(source, []byte("test-reuse-contract=v2"), []byte(marker), 1)
		if bytes.Equal(source, changed) {
			t.Fatal("fixture contract marker was not found")
		}
		if _, err := Load(filename, changed); err == nil {
			t.Fatal("unversioned or historical policy was accepted for current reuse")
		}
	}
}
