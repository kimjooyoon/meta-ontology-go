package toolchainrelease

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func buildBinary(root, output string) error {
	args := []string{"build", "-trimpath", "-buildvcs=true", "-ldflags=-buildid=", "-o", output, "./cmd/gooo"}
	_, err := commandOutput(root, []string{"CGO_ENABLED=0"}, "go", args...)
	return err
}

func commandOutput(root string, overrides []string, name string, args ...string) ([]byte, error) {
	call := exec.Command(name, args...)
	call.Dir = root
	call.Env = mergedEnvironment(os.Environ(), overrides)
	output, err := call.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func mergedEnvironment(base, overrides []string) []string {
	result := append([]string(nil), base...)
	for _, override := range overrides {
		key := strings.SplitN(override, "=", 2)[0] + "="
		filtered := result[:0]
		for _, entry := range result {
			if !strings.HasPrefix(entry, key) {
				filtered = append(filtered, entry)
			}
		}
		result = append(filtered, override)
	}
	return result
}
