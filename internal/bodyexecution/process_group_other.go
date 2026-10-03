//go:build !linux && !darwin

package bodyexecution

import "os/exec"

func bindProcessGroup(cmd *exec.Cmd) {}
