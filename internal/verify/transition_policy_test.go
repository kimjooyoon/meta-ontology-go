package verify

import "testing"

func TestTransitionBranchPolicyIsDeterministic(t *testing.T) {
	if err := CheckPathScope([]string{"scripts/verify.sh", "internal/verify/policy.go"}, []string{".github", "scripts", "internal/verify"}); err != nil {
		t.Fatal(err)
	}
	if err := CheckPathScope([]string{"internal/semantic/graph.go"}, []string{"internal/verify"}); err == nil {
		t.Fatal("core package path crossed CI ownership boundary")
	}
	for _, valid := range [][2]string{{"agent/ci-workflow", "integration"}, {"agent/ci-workflow", "dev"}, {"dev", "main"}} {
		if err := CheckIntegrationPullRequest(valid[0], valid[1]); err != nil {
			t.Fatalf("%s -> %s: %v", valid[0], valid[1], err)
		}
	}
	for _, invalid := range [][2]string{{"agent/ci-workflow", "main"}, {"feature/work", "dev"}, {"dev", "integration"}} {
		if err := CheckIntegrationPullRequest(invalid[0], invalid[1]); err == nil {
			t.Fatalf("invalid branch policy was accepted: %s -> %s", invalid[0], invalid[1])
		}
	}
}
