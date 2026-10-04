package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodyRealizeUsage = "usage: gooo body-realize --source <original.gooo> " +
	"--generation <body-codegen.json> --out <new-directory>"

type bodyRealizationOutput struct {
	bodycodegen.Realization
	GenerationSHA256 string `json:"generation_sha256"`
}

func runBodyRealize(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyRealizeContext(ctx, args, stdout, stderr)
}

func runBodyRealizeContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := map[string]string{"--source": "", "--generation": "", "--out": ""}
	for i := 0; i < len(args); i += 2 {
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, bodyRealizeUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
	}
	for _, value := range flags {
		if value == "" {
			fmt.Fprintln(stderr, bodyRealizeUsage)
			return exitUsage
		}
	}
	return writeBodyRealization(ctx, flags, stdout, stderr)
}

func writeBodyRealization(ctx context.Context, flags map[string]string, stdout, stderr io.Writer) int {
	fail := func(err error) int {
		fmt.Fprintf(stderr, "gooo body-realize: %v\n", err)
		return exitFailure
	}
	source, err := readBodyExecutionFile(flags["--source"], 128<<10)
	if err != nil {
		return fail(err)
	}
	generation, err := readBodyExecutionFile(flags["--generation"], 2<<20)
	if err != nil {
		return fail(err)
	}
	prior, _, err := bodyexecution.DecodeGeneration(generation)
	if err != nil {
		return fail(err)
	}
	realized, err := bodycodegen.RealizeSourceAssembly(ctx, "<body-source>", source, prior)
	if err != nil {
		return fail(err)
	}
	output := bodyRealizationOutput{Realization: realized, GenerationSHA256: fmt.Sprintf("%x", sha256.Sum256(generation))}
	report, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fail(err)
	}
	files := map[string][]byte{"original.gooo": source, "generation.json": generation,
		"realized.gooo": []byte(realized.Source), "realization.json": append(report, '\n')}
	if err := writeRealizationDirectory(ctx, flags["--out"], files); err != nil {
		return fail(err)
	}
	if err := json.NewEncoder(stdout).Encode(output); err != nil {
		return fail(err)
	}
	return exitOK
}

func writeRealizationDirectory(ctx context.Context, directory string, files map[string][]byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Mkdir(directory, 0755); err != nil {
		return fmt.Errorf("output must be a new directory: %w", err)
	}
	for _, name := range []string{"original.gooo", "generation.json", "realized.gooo", "realization.json"} {
		err := ctx.Err()
		if err == nil {
			err = os.WriteFile(filepath.Join(directory, name), files[name], 0644)
		}
		if err != nil {
			return fmt.Errorf("incomplete realization directory %q: %w", directory, err)
		}
	}
	return nil
}
