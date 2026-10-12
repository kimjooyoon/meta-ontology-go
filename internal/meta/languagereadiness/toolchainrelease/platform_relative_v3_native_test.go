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
	if _, err := commandOutput(root, nil, "go", "build", "-o", binary, "./cmd/gooo"); err != nil {
		t.Fatal(err)
	}
	input := BuildInput{Root: root, ExpectedHead: strings.TrimSpace(string(head))}
	if err := smokeRelativeV3(binary, work, input); err != nil {
		t.Fatal(err)
	}
}
