package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodyComposeUsage = "usage: gooo body-compose --source <source.gooo> " +
	"(--cases <cases.json> | --case-series <series.json>) [--repeat <1..16>] " +
	"[--model <model.json>] [--fill-model <model.json>] [--composition <composition.json>] [--go-bin <go1.27.2>] [--out <new-directory>] " +
	"[--assembly-policy <policy.gooo> --policy-activity <name>] [--resume-composition <composition.json>] [--entry <activity>]"

type bodyCompositionOutput struct {
	GeneratedNow   bool                                 `json:"generated_now"`
	Composition    bodyexecution.Composition            `json:"composition"`
	Runtime        bodyexecution.CompositionRuntime     `json:"runtime"`
	RuntimeHistory []bodyexecution.CompositionRuntime   `json:"runtime_history,omitempty"`
	CaseSeries     *bodyexecution.CompositionCaseSeries `json:"case_series,omitempty"`
}

func runBodyCompose(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodyComposeContext(ctx, args, stdout, stderr)
}

func runBodyComposeContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprintln(stderr, bodyComposeUsage)
		return exitOK
	}
	flags := map[string]string{"--source": "", "--cases": "", "--case-series": "", "--repeat": "", "--model": "", "--fill-model": "", "--composition": "", "--go-bin": "", "--out": ""}
	flags["--assembly-policy"], flags["--policy-activity"] = "", ""
	flags["--resume-composition"] = ""
	flags["--entry"] = ""
	for i := 0; i < len(args); i += 2 {
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, bodyComposeUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
	}
	if flags["--source"] == "" || (flags["--cases"] == "") == (flags["--case-series"] == "") ||
		((flags["--model"] != "" || flags["--fill-model"] != "" || flags["--assembly-policy"] != "") && flags["--composition"] != "") ||
		((flags["--assembly-policy"] == "") != (flags["--policy-activity"] == "")) {
		fmt.Fprintln(stderr, bodyComposeUsage)
		return exitUsage
	}
	if value := flags["--repeat"]; value != "" {
		count, err := strconv.Atoi(value)
		if err != nil || count < 1 || count > 16 {
			fmt.Fprintln(stderr, bodyComposeUsage)
			return exitUsage
		}
	}
	if flags["--resume-composition"] != "" && (flags["--assembly-policy"] == "" || flags["--composition"] != "" ||
		flags["--model"] != "" || flags["--fill-model"] != "") {
		fmt.Fprintln(stderr, bodyComposeUsage)
		return exitUsage
	}
	if flags["--entry"] != "" && (flags["--composition"] != "" || flags["--resume-composition"] != "") {
		fmt.Fprintln(stderr, bodyComposeUsage)
		return exitUsage
	}
	return executeBodyComposition(ctx, flags, stdout, stderr)
}

func executeBodyComposition(ctx context.Context, flags map[string]string, stdout, stderr io.Writer) int {
	fail := func(err error) int { fmt.Fprintf(stderr, "gooo body-compose: %v\n", err); return exitFailure }
	if directory := flags["--out"]; directory != "" {
		if _, err := os.Lstat(directory); !os.IsNotExist(err) {
			return fail(fmt.Errorf("output must be a new directory"))
		}
	}
	source, err := readBodyExecutionFile(flags["--source"], 128<<10)
	if err != nil {
		return fail(err)
	}
	suites, cases, series, err := readCompositionSuites(flags)
	if err != nil {
		return fail(err)
	}
	repeat := 1
	if flags["--repeat"] != "" {
		repeat, _ = strconv.Atoi(flags["--repeat"])
	}
	if len(suites)*repeat > 16 {
		return fail(fmt.Errorf("composition allows at most 16 executions per request"))
	}
	if series != nil && flags["--composition"] == "" && flags["--resume-composition"] == "" {
		if err := bodyexecution.ValidateCompositionSuitesForEntry(ctx, flags["--source"], source, suites, flags["--entry"]); err != nil {
			return fail(err)
		}
	}
	output := bodyCompositionOutput{GeneratedNow: flags["--composition"] == "", CaseSeries: series}
	output.Composition, err = buildOrReadBodyComposition(ctx, flags, source, suites[0])
	if err == nil && series != nil && (flags["--composition"] != "" || flags["--resume-composition"] != "") {
		err = bodyexecution.ValidateCompositionSuitesForEntry(ctx, flags["--source"], source, suites, output.Composition.Plan.EntryActivity)
	}
	if err == nil {
		if series != nil || repeat > 1 {
			output.RuntimeHistory, err = executeCompositionHistory(ctx, flags, source, output.Composition, suites, repeat)
			if len(output.RuntimeHistory) > 0 {
				output.Runtime = output.RuntimeHistory[len(output.RuntimeHistory)-1]
			}
		} else {
			output.Runtime, err = bodyexecution.ExecuteComposition(ctx, flags["--source"], source, output.Composition, suites[0], flags["--go-bin"])
		}
	}
	if directory := flags["--out"]; directory != "" {
		if writeErr := writeCompositionOutput(directory, source, cases, output); writeErr != nil {
			return fail(writeErr)
		}
	}
	if encodeErr := json.NewEncoder(stdout).Encode(output); encodeErr != nil {
		return fail(encodeErr)
	}
	if err != nil {
		return fail(err)
	}
	return exitOK
}

func writeCompositionOutput(directory string, source, cases []byte, output bodyCompositionOutput) error {
	if err := os.Mkdir(directory, 0755); err != nil {
		return fmt.Errorf("output must be a new directory: %w", err)
	}
	composition, err := json.MarshalIndent(output.Composition, "", "  ")
	if err != nil {
		return err
	}
	runtime, err := json.MarshalIndent(output.Runtime, "", "  ")
	if err != nil {
		return err
	}
	files := []struct {
		name string
		data []byte
	}{
		{"original.gooo", source}, {"cases.json", cases}, {"composition.json", append(composition, '\n')},
		{"runtime.json", append(runtime, '\n')}, {"realized.gooo", []byte(output.Composition.GoooSource)},
		{"generated.go", []byte(output.Composition.Source)},
		{"main.go", []byte(output.Composition.Driver)},
		{"go.mod", []byte("module gooo.observed.composition\n\ngo 1.27.2\n")},
	}
	if len(output.RuntimeHistory) > 0 {
		data, err := json.MarshalIndent(output.RuntimeHistory, "", "  ")
		if err != nil {
			return err
		}
		files = append(files, struct {
			name string
			data []byte
		}{"runtime-history.json", append(data, '\n')})
	}
	if output.CaseSeries != nil {
		data, err := json.MarshalIndent(output.CaseSeries, "", "  ")
		if err != nil {
			return err
		}
		files = append(files, struct {
			name string
			data []byte
		}{"case-series.json", append(data, '\n')})
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(directory, file.name), file.data, 0644); err != nil {
			return err
		}
	}
	return nil
}
