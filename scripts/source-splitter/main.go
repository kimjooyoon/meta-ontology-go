package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "source-splitter:", err)
		os.Exit(1)
	}
}

func repositorySHA(root string) (string, error) {
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("read repository SHA: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
