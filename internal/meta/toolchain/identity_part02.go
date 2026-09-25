package toolchain

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Identity binds the declared module toolchain to three runtime observations.
func Identity(root string) (string, error) {
	requirement, err := ReadRequirement(root)
	if err != nil {
		return "", err
	}
	goVersion, err := command(root, "go", "version")
	if err != nil {
		return "", err
	}
	goEnv, err := command(root, "go", "env", "GOVERSION", "GOROOT", "GOOS", "GOARCH")
	if err != nil {
		return "", err
	}
	observed := runtime.Version()
	if observed != requirement.Toolchain || !strings.Contains(goVersion, requirement.Toolchain) || !strings.HasPrefix(goEnv, requirement.Toolchain+"\n") {
		return "", fmt.Errorf("toolchain mismatch: declared=%q runtime=%q go=%q env=%q", requirement.Toolchain, observed, strings.TrimSpace(goVersion), strings.TrimSpace(goEnv))
	}
	return observed + "\n" + strings.TrimSpace(goVersion) + "\n" + strings.TrimSpace(goEnv), nil
}

func command(root, name string, args ...string) (string, error) {
	call := exec.Command(name, args...)
	call.Dir = root
	output, err := call.Output()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(output), nil
}
