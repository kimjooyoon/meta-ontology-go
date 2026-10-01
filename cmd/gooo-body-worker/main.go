// gooo-body-worker is an opt-in native NDJSON construction experiment.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodypathstream"
)

func main() { os.Exit(run()) }

func run() int {
	model := flag.String("model", "", "optional explicit local structural model.json; empty is deterministic")
	workers := flag.Int("workers", 1, "bounded native construction workers (1..8)")
	flag.Parse()
	if flag.NArg() != 0 || *workers < 1 || *workers > bodypathstream.MaxWorkers {
		fmt.Fprintln(os.Stderr, "usage: gooo-body-worker [--model model.json] [--workers 1..8]")
		return 2
	}
	generator, err := bodycodegen.NewTypedPathGenerator(*model)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err = json.NewEncoder(os.Stderr).Encode(generator.Info()); err != nil {
		return 1
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err = bodypathstream.Run(ctx, generator, os.Stdin, os.Stdout, *workers); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
