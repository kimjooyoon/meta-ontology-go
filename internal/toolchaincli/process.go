package toolchaincli

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

const invocationTimeout = 5 * time.Second
const invocationOutputWait = time.Second

func invoke(executable, root string, arguments []string) (Observation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), invocationTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, executable, arguments...)
	command.Dir, command.WaitDelay = root, invocationOutputWait
	command.Env = []string{"LANG=C", "LC_ALL=C", "TZ=UTC", "NO_COLOR=1"}
	output, err := runCapturedCommand(ctx, command)
	result := Observation{Arguments: append([]string(nil), arguments...), ExitCode: 0,
		Stdout: output.Stdout, Stderr: output.Stderr}
	var resourceErr error
	result.PeakRSSKiB, resourceErr = peakRSSKiB(command.ProcessState)
	if ctx.Err() != nil {
		result.ExitCode, result.Failure = -1, "INVOCATION_DEADLINE"
	} else if output.Overflow {
		result.ExitCode, result.Failure = -1, "INVOCATION_OUTPUT_LIMIT"
	} else if err != nil && command.ProcessState == nil {
		result.ExitCode = -1
		return result, errors.Join(err, output.Err)
	} else if output.Err != nil {
		result.ExitCode, result.Failure = command.ProcessState.ExitCode(), "INVOCATION_OUTPUT_INCOMPLETE"
		return result, errors.Join(err, output.Err)
	} else if resourceErr != nil {
		result.ExitCode, result.Failure = -1, "INVOCATION_RESOURCE_UNKNOWN"
	} else if err != nil {
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return result, err
		}
	}
	if output.Err != nil {
		return result, errors.Join(err, output.Err)
	}
	return result, nil
}
