package toolchaincli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type invocationOutput struct {
	Stdout, Stderr string
	Overflow       bool
	Err            error
}

// CaptureState describes retained bytes separately from reaching stream EOF.
// It is an internal error detail; it does not change readiness report schemas.
type CaptureState struct {
	EOF, DeadlineCloseRequested, LimitExceeded bool
	RetainedBytes                              int
}

// OutputCaptureError preserves both stream states when capture is incomplete.
type OutputCaptureError struct {
	Stdout, Stderr CaptureState
	Cause          error
}

func (err *OutputCaptureError) Error() string {
	if err.Cause == nil {
		return "CLI output capture incomplete"
	}
	return "CLI output capture incomplete: " + err.Cause.Error()
}

func (err *OutputCaptureError) Unwrap() error { return err.Cause }

type invocationStream struct {
	reader, writer *os.File
	buffer         cappedBuffer
	done           chan error
}

func runCapturedCommand(ctx context.Context, command *exec.Cmd) (invocationOutput, error) {
	stdout, err := openInvocationStream()
	if err != nil {
		return invocationOutput{}, fmt.Errorf("open stdout capture: %w", err)
	}
	defer stdout.close()
	stderr, err := openInvocationStream()
	if err != nil {
		return invocationOutput{}, fmt.Errorf("open stderr capture: %w", err)
	}
	defer stderr.close()
	command.Stdout, command.Stderr = stdout.writer, stderr.writer
	stdout.start()
	stderr.start()
	runErr := command.Run()
	stdoutCloseErr, stderrCloseErr := stdout.writer.Close(), stderr.writer.Close()
	drainContext, cancel := context.WithTimeout(ctx, invocationOutputWait)
	defer cancel()
	stdoutText, stdoutState, stdoutErr := stdout.finish(drainContext)
	stderrText, stderrState, stderrErr := stderr.finish(drainContext)
	result := invocationOutput{Stdout: stdoutText, Stderr: stderrText,
		Overflow: stdout.buffer.overflow || stderr.buffer.overflow}
	cause := errors.Join(captureError("stdout", errors.Join(stdoutCloseErr, stdoutErr)),
		captureError("stderr", errors.Join(stderrCloseErr, stderrErr)))
	if cause != nil {
		result.Err = &OutputCaptureError{Stdout: stdoutState, Stderr: stderrState, Cause: cause}
	}
	return result, runErr
}

func openInvocationStream() (*invocationStream, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &invocationStream{reader: reader, writer: writer, done: make(chan error, 1)}, nil
}

func (stream *invocationStream) start() {
	go func() {
		_, err := io.Copy(&stream.buffer, stream.reader)
		stream.done <- err
	}()
}

func (stream *invocationStream) finish(ctx context.Context) (string, CaptureState, error) {
	select {
	case err := <-stream.done:
		return stream.result(err, false, nil)
	default:
	}
	select {
	case err := <-stream.done:
		return stream.result(err, false, nil)
	case <-ctx.Done():
		closeErr := stream.reader.Close()
		err := <-stream.done
		return stream.result(err, true, errors.Join(ctx.Err(), closeErr))
	}
}

func (stream *invocationStream) result(
	copyErr error, requestedClose bool, closeErr error,
) (string, CaptureState, error) {
	state := CaptureState{EOF: copyErr == nil, DeadlineCloseRequested: requestedClose,
		LimitExceeded: stream.buffer.overflow, RetainedBytes: len(stream.buffer.data)}
	return stream.buffer.String(), state, errors.Join(copyErr, closeErr)
}

func (stream *invocationStream) close() {
	_ = stream.reader.Close()
	_ = stream.writer.Close()
}

func captureError(stream string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s capture incomplete: %w", stream, err)
}
