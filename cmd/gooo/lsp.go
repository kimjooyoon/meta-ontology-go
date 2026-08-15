package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

const lspUsage = "usage: gooo lsp"

func runLSP(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, lspUsage)
		return exitUsage
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runLSPContext(ctx, input, stdout, stderr)
}

func runLSPContext(ctx context.Context, input io.Reader, stdout, stderr io.Writer) int {
	if err := lsp.NewServer().ServeContext(ctx, input, stdout); err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Fprintf(stderr, "gooo: lsp: %v\n", err)
		}
		return exitFailure
	}
	return exitOK
}
