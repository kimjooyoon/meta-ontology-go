package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodypathstream"
)

func runBodyPathStream(args []string, input io.Reader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return bodypathstream.RunCommand(ctx, "gooo body-path-stream", args, input, stdout, stderr)
}
