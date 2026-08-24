package main

import (
	"errors"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type runSourceOptions struct {
	filename string
	entry    string
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
		default:
			if strings.HasPrefix(args[index], "-") || options.filename != "" {
				return runSourceOptions{}, errors.New(runSourceUsage)
			}
			options.filename = args[index]
		}
	}
	if options.filename == "" || strings.TrimSpace(options.entry) == "" {
		return runSourceOptions{}, errors.New(runSourceUsage)
	}
	return options, nil
}

func sourceexecutionSpan() syntax.Span { return syntax.Span{} }
