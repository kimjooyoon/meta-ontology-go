package safeworkbinding

import (
	"testing"
)

func TestBindingDigest_GovernedFieldMutations(t *testing.T) {
	mutations := []struct {
		name   string
		mutate func(*SafeWorkBinding)
	}{
		{
			"schema",
			func(binding *SafeWorkBinding) { binding.Schema += "-v2" },
		},
		{
			"task",
			func(binding *SafeWorkBinding) { binding.TaskID += "-v2" },
		},
		{
			"path",
			func(binding *SafeWorkBinding) { binding.PathID += "-v2" },
		},
		{
			"obligation",
			func(binding *SafeWorkBinding) { binding.ObligationID += "-v2" },
		},
		{
			"source",
			func(binding *SafeWorkBinding) { binding.SourceSnapshotDigest += "-v2" },
		},
		{
			"semantic",
			func(binding *SafeWorkBinding) { binding.SemanticSnapshotDigest += "-v2" },
		},
		{
			"policy",
			func(binding *SafeWorkBinding) { binding.PolicyDigest += "-v2" },
		},
		{
			"registry",
			func(binding *SafeWorkBinding) { binding.RegistryDigest += "-v2" },
		},
		{
			"toolchain",
			func(binding *SafeWorkBinding) { binding.ToolchainOptionsDigest += "-v2" },
		},
	}
	base := bindingDigest(baseBindingForDigest())
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			binding := baseBindingForDigest()
			mutation.mutate(&binding)
			check(t, bindingDigest(binding) != base, "governed mutation")
		})
	}
}
