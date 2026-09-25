package verify

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGuardianJavaScriptSuite(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "scripts/ci-proof/guardian_test.js")
	command.Dir = root
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("guardian JavaScript suite timed out: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("guardian JavaScript suite failed: %v\n%s", err, strings.TrimSpace(string(output)))
	}
	if !strings.Contains(string(output), "guardian tests passed") {
		t.Fatalf("guardian JavaScript suite did not emit its success marker: %s", strings.TrimSpace(string(output)))
	}
}
