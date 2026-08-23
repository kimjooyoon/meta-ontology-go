package toolchaincli

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

const invocationTimeout = 5 * time.Second

func invoke(executable, root string, arguments []string) (Observation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), invocationTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, executable, arguments...)
	command.Dir = root
	command.Env = []string{"LANG=C", "LC_ALL=C", "TZ=UTC", "NO_COLOR=1"}
	stdout, stderr := &cappedBuffer{}, &cappedBuffer{}
	command.Stdout, command.Stderr = stdout, stderr
	err := command.Run()
	result := Observation{Arguments: append([]string(nil), arguments...), ExitCode: 0,
		Stdout: stdout.String(), Stderr: stderr.String()}
	if ctx.Err() != nil {
		result.ExitCode, result.Failure = -1, "INVOCATION_DEADLINE"
	} else if stdout.overflow || stderr.overflow {
		result.ExitCode, result.Failure = -1, "INVOCATION_OUTPUT_LIMIT"
	} else if err != nil {
		if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return result, err
		}
	}
	return result, nil
}
