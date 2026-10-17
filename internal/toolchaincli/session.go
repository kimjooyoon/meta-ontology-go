package toolchaincli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (session Session) BinaryDigest() (string, error) {
	if !filepath.IsAbs(session.Executable) {
		return "", fmt.Errorf("toolchain executable must be absolute")
	}
	info, err := os.Stat(session.Executable)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("toolchain executable is not regular")
	}
	return digestFile(session.Executable)
}

func (session Session) Invoke(arguments []string) (Observation, error) {
	if !filepath.IsAbs(session.Root) {
		return Observation{}, fmt.Errorf("repository root must be absolute")
	}
	before, err := snapshot(session.Root)
	if err != nil {
		return Observation{}, err
	}
	result, invokeErr := invoke(session.Executable, session.Root, arguments)
	after, err := snapshot(session.Root)
	return completeInvocation(result, before, after, invokeErr, err)
}

func completeInvocation(
	result Observation, before, after treeSnapshot, invokeErr, snapshotErr error,
) (Observation, error) {
	result.TreeBeforeDigest = before.Digest
	if snapshotErr != nil {
		return result, errors.Join(invokeErr, snapshotErr)
	}
	result.TreeAfterDigest = after.Digest
	result.RepositoryWrites = changedFiles(before, after)
	return result, invokeErr
}
