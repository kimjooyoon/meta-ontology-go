package toolchainrelease

import (
	"fmt"
	"strings"
)

func repositoryState(root string) (string, error) {
	output, err := commandOutput(root, nil, "git", "status", "--porcelain=v1", "--untracked-files=all")
	return strings.TrimSpace(string(output)), err
}

func requireRepositoryHead(root, expected string) error {
	output, err := commandOutput(root, nil, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(output)) != expected {
		return fmt.Errorf("TOOLCHAIN_RELEASE_CHECKOUT_HEAD_MISMATCH")
	}
	return nil
}
