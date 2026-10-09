package toolchainrelease

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPlatformCommandKeepsFailedOutput(t *testing.T) {
	const output = `{"stage":"EXECUTE_1","completed":false}`
	if os.Getenv("GOOO_PLATFORM_FAILURE_HELPER") == "1" {
		fmt.Print(output)
		os.Exit(9)
	}
	raw, err := commandOutput(t.TempDir(), []string{"GOOO_PLATFORM_FAILURE_HELPER=1"},
		os.Args[0], "-test.run=^TestPlatformCommandKeepsFailedOutput$")
	if err == nil || !bytes.Equal(raw, []byte(output)) || !strings.Contains(err.Error(), output) {
		t.Fatalf("failed command output was lost: output=%q error=%v", raw, err)
	}
}
