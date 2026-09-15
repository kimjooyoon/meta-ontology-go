package exactsha

import (
	"fmt"
	"os/exec"
	"strings"
)

func Verify(root, expected string) error {
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read repository SHA: %w", err)
	}
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		return fmt.Errorf("repository SHA %s does not match evidence %s", actual, expected)
	}
	return nil
}
