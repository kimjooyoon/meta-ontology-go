package adkcheck

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// CI supplies a binary built from the same revision as the optional bridge.
// A plain module test remains independent of the compiler checkout.
func TestCurrentCLIParity(t *testing.T) {
	binary := os.Getenv("GOOO_ADK_TEST_BINARY")
	if binary == "" {
		t.Skip("set GOOO_ADK_TEST_BINARY to a current Gooo binary")
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	valid := "package billing\nnamespace billing\nentity Order id \"billing://order\"\n"
	invalid := valid + "activity Pay(Missing) -> Order\n"
	for name, source := range map[string]string{"valid.gooo": valid, "invalid.gooo": invalid} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name     string
		path     string
		wantExit int
	}{
		{"valid", "valid.gooo", 0},
		{"semantic error", "invalid.gooo", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := check(context.Background(), binary, root, 15*time.Second, CheckArgs{Path: tc.path})
			if err != nil {
				t.Fatal(err)
			}
			if result.ExitCode != tc.wantExit || !json.Valid([]byte(result.ReportJSON)) {
				t.Fatalf("exit=%d, valid JSON=%t; expected exit=%d", result.ExitCode, json.Valid([]byte(result.ReportJSON)), tc.wantExit)
			}
		})
	}
	t.Run("discovery", func(t *testing.T) {
		result, err := discover(context.Background(), binary, root, 15*time.Second, DiscoverArgs{
			Path: "valid.gooo", Query: "billing order",
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.ExitCode != 0 || !json.Valid([]byte(result.ReportJSON)) {
			t.Fatalf("exit=%d, valid JSON=%t", result.ExitCode, json.Valid([]byte(result.ReportJSON)))
		}
	})
}
