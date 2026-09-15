package main

import (
	"errors"
	"strings"
)

const emitUsage = "usage: gooo emit --kind <kind> --entry <activity> <package-directory>"

type emitOptions struct {
	directory string
	entry     string
	kind      string
}

func parseEmitArguments(args []string) (emitOptions, error) {
	var options emitOptions
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--entry":
			if options.entry != "" || index+1 >= len(args) {
				return emitOptions{}, errors.New(emitUsage)
			}
			index++
			options.entry = args[index]
		case "--kind":
			if options.kind != "" || index+1 >= len(args) {
				return emitOptions{}, errors.New(emitUsage)
			}
			index++
			options.kind = args[index]
		default:
			if strings.HasPrefix(args[index], "-") || options.directory != "" {
				return emitOptions{}, errors.New(emitUsage)
			}
			options.directory = args[index]
		}
	}
	if options.directory == "" || strings.TrimSpace(options.entry) == "" || strings.TrimSpace(options.kind) == "" {
		return emitOptions{}, errors.New(emitUsage)
	}
	return options, nil
}
