package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func compositionInputModes(flags map[string]string) (count int) {
	for _, flag := range []string{"--cases", "--inputs", "--case-series"} {
		if flags[flag] != "" {
			count++
		}
	}
	return count
}

func readCompositionSuites(flags map[string]string) ([]bodyexecution.CompositionCases, []byte,
	*bodyexecution.CompositionCaseSeries, error) {
	if flags["--case-series"] == "" {
		path, decode := flags["--cases"], bodyexecution.DecodeCompositionCases
		if flags["--inputs"] != "" {
			path, decode = flags["--inputs"], bodyexecution.DecodeCompositionInputs
		}
		raw, err := readBodyExecutionFile(path, 32<<10)
		if err != nil {
			return nil, nil, nil, err
		}
		suite, err := decode(raw)
		return []bodyexecution.CompositionCases{suite}, raw, nil, err
	}
	raw, err := readBodyExecutionFile(flags["--case-series"], 512<<10)
	if err != nil {
		return nil, nil, nil, err
	}
	series, err := bodyexecution.DecodeCompositionCaseSeries(raw)
	if err != nil {
		return nil, nil, nil, err
	}
	first, err := json.MarshalIndent(series.Suites[0], "", "  ")
	return series.Suites, append(first, '\n'), &series, err
}

func executeCompositionHistory(ctx context.Context, flags map[string]string, source []byte,
	prior bodyexecution.Composition, suites []bodyexecution.CompositionCases, repeat int) (history []bodyexecution.CompositionRuntime, err error) {
	executor := bodyexecution.NewExecutor()
	defer func() { err = errors.Join(err, executor.Close()) }()
	history = make([]bodyexecution.CompositionRuntime, 0, len(suites)*repeat)
	for range repeat {
		for _, suite := range suites {
			r, runErr := executor.ExecuteComposition(ctx, flags["--source"], source, prior, suite, flags["--go-bin"])
			history = append(history, r)
			if runErr != nil {
				return history, runErr
			}
		}
	}
	return history, nil
}
