package publictestreuse

import (
	"bytes"
	"os"
	"testing"
)

func TestPolicyRequiresObservedPackageAndToolContract(t *testing.T) {
	filename := "../../../examples/self-improvement-discovery/project.gooo"
	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Load(filename, source)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Schema != PolicySchema || len(policy.Bindings) != 13 ||
		policy.Bindings[10] != "execution-package-inputs" || policy.Bindings[11] != "execution-tool-inputs" {
		t.Fatal("current policy omitted observed execution package or tool inputs")
	}
	for _, marker := range []string{"test-reuse-contract=v1", "test-reuse-contract=v2", "test-reuse-contract=v3", "test-reuse-contract="} {
		changed := bytes.Replace(source, []byte("test-reuse-contract=v4"), []byte(marker), 1)
		if bytes.Equal(source, changed) {
			t.Fatal("fixture contract marker was not found")
		}
		if _, err := Load(filename, changed); err == nil {
			t.Fatal("unversioned or historical policy was accepted for current reuse")
		}
	}
	for _, binding := range []string{"execution-package-inputs", "execution-tool-inputs"} {
		changed := bytes.Replace(source, []byte(binding+","), nil, 1)
		if bytes.Equal(source, changed) {
			t.Fatal("fixture binding was not found", binding)
		}
		if _, err := Load(filename, changed); err == nil {
			t.Fatal("source with an omitted execution binding was accepted", binding)
		}
	}
}
