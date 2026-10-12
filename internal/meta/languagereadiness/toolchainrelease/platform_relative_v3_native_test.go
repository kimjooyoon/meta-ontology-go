package toolchainrelease

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This fresh profile executes only in native CI. Its zero-weight artifact is
// independent of every fitted study and earlier native measurement.
func TestRelativeV3ReleaseProfileNativeCI(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("new release-profile native execution is CI-only")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	head, err := commandOutput(root, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	binary := filepath.Join(work, "gooo")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	state, err := repositoryState(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := buildBinary(root, binary); err != nil {
		t.Fatal(err)
	}
	target, ok := expectedTarget(runtime.GOOS + "-" + runtime.GOARCH)
	if !ok {
		t.Fatal("native release target is unregistered")
	}
	input := BuildInput{Root: root, ExpectedHead: strings.TrimSpace(string(head)), Target: target}
	err = smokeRelativeV3(binary, work, input)
	if state != "" {
		if err == nil || err.Error() != "TOOLCHAIN_RELEASE_DIRTY_BUILD" {
			t.Fatalf("rewritten source must be rejected before inference: %v", err)
		}
		entries, readErr := os.ReadDir(work)
		if readErr != nil || len(entries) != 1 || entries[0].Name() != filepath.Base(binary) {
			t.Fatalf("rejected source created profile artifacts: %v, %v", entries, readErr)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
}
