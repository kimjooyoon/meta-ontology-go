package main

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type runSourceOptions struct {
	iterations  int
	filename    string
	entry       string
	input       string
	record      bool
	runtimePlan string
}

func parseRunSourceArguments(args []string) (runSourceOptions, error) {
	var options runSourceOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--entry":
			if options.entry != "" || index+1 >= len(args) {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			index++
			options.entry = args[index]
		case "--iterations":
			if options.iterations != 0 || index+1 >= len(args) {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			index++
			value, err := strconv.Atoi(args[index])
			if err != nil || value < 1 {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			options.iterations = value
		case "--input", "--record-input":
			if options.input != "" || index+1 >= len(args) {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			options.record = args[index] == "--record-input"
			index++
			options.input = args[index]
			if options.record && options.input == "" {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
		case "--runtime-plan":
			if options.runtimePlan != "" || index+1 >= len(args) {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			index++
			options.runtimePlan = args[index]
			if options.runtimePlan == "" {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
		default:
			if strings.HasPrefix(args[index], "-") || options.filename != "" {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			options.filename = args[index]
		}
	}
	if options.filename == "" || strings.TrimSpace(options.entry) == "" ||
		(options.iterations > 0 && (options.input == "" || options.record)) ||
		(options.runtimePlan != "" && (options.input == "" || options.record)) {
		return runSourceOptions{}, errors.New(runSourceUsage)
	}
	return options, nil
}

func sourceexecutionSpan() syntax.Span { return syntax.Span{} }
