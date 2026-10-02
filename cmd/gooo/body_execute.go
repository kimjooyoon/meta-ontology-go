package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodyExecuteUsage = "usage: gooo body-execute --source <original.gooo> --path-plan <plan.json> " +
	"--generation <body-codegen.json> --cases <cases.json> [--go-bin <go1.27.1>]"

func runBodyExecute(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyExecuteContext(ctx, args, stdout, stderr)
}

func runBodyExecuteContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := map[string]string{"--source": "", "--path-plan": "", "--generation": "", "--cases": "", "--go-bin": ""}
	for i := 0; i < len(args); i += 2 {
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, bodyExecuteUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
	}
	data := map[string][]byte{}
	for _, field := range []struct {
		name  string
		limit int64
	}{
		{"--source", 128 << 10}, {"--path-plan", 256 << 10}, {"--generation", 2 << 20}, {"--cases", 32 << 10},
	} {
		if flags[field.name] == "" {
			fmt.Fprintln(stderr, bodyExecuteUsage)
			return exitUsage
		}
		raw, err := readBodyExecutionFile(flags[field.name], field.limit)
		if err != nil {
			fmt.Fprintf(stderr, "gooo body-execute: %s: %v\n", field.name, err)
			return exitFailure
		}
		data[field.name] = raw
	}
	prior, parent, err := bodyexecution.DecodeGeneration(data["--generation"])
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-execute: generation: %v\n", err)
		return exitFailure
	}
	plan, err := bodyexecution.DecodePlan(data["--path-plan"])
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-execute: plan: %v\n", err)
		return exitFailure
	}
	cases, err := bodyexecution.DecodeCases(data["--cases"])
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-execute: cases: %v\n", err)
		return exitFailure
	}
	result, runErr := bodyexecution.Execute(ctx, "<body-source>", data["--source"], plan, prior, parent, cases, flags["--go-bin"])
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	if runErr != nil {
		fmt.Fprintf(stderr, "gooo body-execute: %v\n", runErr)
		return exitFailure
	}
	return exitOK
}

func readBodyExecutionFile(path string, limit int64) ([]byte, error) {
	before, err := os.Stat(path)
	if err != nil || !before.Mode().IsRegular() {
		return nil, fmt.Errorf("input must be a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open input")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > limit {
		return nil, fmt.Errorf("input must be a nonempty regular file within %d bytes", limit)
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || len(raw) == 0 || int64(len(raw)) > limit {
		return nil, fmt.Errorf("input read exceeds bound or failed")
	}
	return raw, nil
}
