package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodypathstream"
)

func runBodyPathFiles(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return bodypathstream.RunFilesCommand(ctx, "gooo body-path-run", args, stdout, stderr)
}
