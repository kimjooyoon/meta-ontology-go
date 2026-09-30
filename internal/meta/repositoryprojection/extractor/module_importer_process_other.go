//go:build !darwin && !linux

package extractor

import "os/exec"

func configureModuleListProcess(command *exec.Cmd) {
	command.WaitDelay = moduleListWaitDelay
}
