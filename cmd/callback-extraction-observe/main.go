package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("callback-extraction-observe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root := flags.String("root", ".", "source repository root (read only)")
	file := flags.String("file", "", "repository-relative Go source path")
	subject := flags.String("subject", "", "source test subject, e.g. func:TestExample")
	timeout := flags.Duration("timeout", 4*time.Minute, "positive CI observation deadline")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *file == "" || *subject == "" || *timeout <= 0 {
		fmt.Fprintln(stderr, "file, subject and a positive timeout are required; positional arguments are forbidden")
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	observation, observeError := extractor.ObserveCallbackExtraction(ctx, *root, *file, *subject)
	if err := json.NewEncoder(stdout).Encode(observation); err != nil {
		fmt.Fprintln(stderr, "write observation:", err)
		return 1
	}
	if observeError != nil {
		fmt.Fprintln(stderr, observeError)
		return 1
	}
	return 0
}
