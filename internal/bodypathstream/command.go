package bodypathstream

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// RunCommand serves the same protocol for the integrated CLI and standalone
// worker. Transport closers must unblock pending I/O when cancellation closes
// them. Normal EOF leaves caller-owned streams open. Exit codes are 0 for EOF
// or help, 1 for setup/transport failure, and 2 for invalid arguments.
func RunCommand(ctx context.Context, name string, args []string, input io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	model := flags.String("model", "", "explicit local model.json; omit for deterministic construction")
	workers := flags.Int("workers", 1, "bounded construction workers (1..8)")
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s [--model model.json] [--workers 1..8] < requests.jsonl\n", name)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *workers < 1 || *workers > MaxWorkers {
		flags.Usage()
		return 2
	}
	reader, readOK := input.(io.ReadCloser)
	writer, writeOK := stdout.(io.WriteCloser)
	if !readOK || !writeOK {
		fmt.Fprintf(stderr, "%s: input and output must support Close for cancellation\n", name)
		return 1
	}
	if err := runCommandStream(ctx, *model, reader, writer, stderr, *workers); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", name, err)
		return 1
	}
	return 0
}

func runCommandStream(ctx context.Context, model string, input io.ReadCloser,
	stdout io.WriteCloser, stderr io.Writer, workers int) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	generator, err := bodycodegen.NewTypedPathGenerator(model)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(stderr).Encode(generator.Info()); err != nil {
		return fmt.Errorf("write setup record: %w", err)
	}
	return Run(ctx, generator, input, stdout, workers)
}
