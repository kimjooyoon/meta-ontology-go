package moduleboundary

import (
	"fmt"
	"go/version"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LocalToolchainSupported reports whether the fixed local Go toolchain meets
// the owning module's minimum go directive. A missing manifest preserves the
// legacy repository-root fallback.
func LocalToolchainSupported(owner string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(owner, "go.mod"))
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "go" {
			continue
		}
		required := "go" + fields[1]
		if !version.IsValid(required) {
			return false, fmt.Errorf("invalid Go requirement in module %s", owner)
		}
		current := runtime.Version()
		if !version.IsValid(current) {
			return false, fmt.Errorf("unrecognized local Go toolchain %q", current)
		}
		return version.Compare(current, required) >= 0, nil
	}
	return true, nil // Go treats an omitted go line as an older implicit version.
}
