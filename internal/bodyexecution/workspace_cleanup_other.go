//go:build !windows

package bodyexecution

func retryableWorkspaceRemoval(error) bool { return false }
