package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

const packageSessionUsage = "usage: gooo package session (one JSON request per line; operations execute, replay, close)"
const packageSessionRequestSchema = "gooo/workspace-session-request/v1"
const packageSessionResponseSchema = "gooo/workspace-session-response/v1"
const packageSessionMaxRequestBytes = 64 << 10

type packageSessionRequest struct {
	Schema    string   `json:"schema"`
	ID        string   `json:"id"`
	Operation string   `json:"operation"`
	Arguments []string `json:"arguments,omitempty"`
}

type packageSessionResponse struct {
	Schema   string                   `json:"schema"`
	ID       string                   `json:"id"`
	ExitCode int                      `json:"exit_code"`
	Receipt  *packageExecutionReceipt `json:"receipt,omitempty"`
	Closed   bool                     `json:"closed,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

func runPackageSession(args []string, input io.Reader, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprintln(stderr, packageSessionUsage)
		return exitOK
	}
	if len(args) != 0 {
		fmt.Fprintln(stderr, packageSessionUsage)
		return exitUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return servePackageSession(ctx, input, reader, stdout, stderr)
}

func servePackageSession(ctx context.Context, input io.Reader, reader SourceReader, stdout, stderr io.Writer) int {
	owner := workspaceexecution.NewExecutor()
	stop := context.AfterFunc(ctx, func() {
		_ = owner.Close()
		for _, stream := range []any{input, stdout, stderr} {
			if closer, ok := stream.(io.Closer); ok {
				_ = closer.Close()
			}
		}
	})
	defer stop()
	code := packageSessionLoop(ctx, input, reader, stdout, stderr, owner)
	if err := owner.Close(); err != nil {
		fmt.Fprintln(stderr, "gooo package session: cleanup:", err)
		return exitFailure
	}
	return code
}

func packageSessionLoop(ctx context.Context, input io.Reader, reader SourceReader,
	stdout, stderr io.Writer, owner *workspaceexecution.Executor) int {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), packageSessionMaxRequestBytes)
	encoder := json.NewEncoder(stdout)
	for scanner.Scan() {
		if ctx.Err() != nil {
			break
		}
		response := handlePackageSessionLine(ctx, scanner.Bytes(), reader, owner, stderr)
		if err := encoder.Encode(response); err != nil {
			fmt.Fprintln(stderr, "gooo package session: write response:", err)
			return exitFailure
		}
		if response.Closed {
			return response.ExitCode
		}
	}
	if err := errors.Join(scanner.Err(), ctx.Err()); err != nil {
		fmt.Fprintln(stderr, "gooo package session:", err)
		return exitFailure
	}
	return exitOK
}
