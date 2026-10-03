// gooo-body-worker is an opt-in native NDJSON construction experiment.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodypathstream"
)

func main() { os.Exit(run()) }

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return bodypathstream.RunCommand(ctx, "gooo-body-worker", os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
}
