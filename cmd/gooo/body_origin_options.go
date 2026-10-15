package main

import (
	"fmt"
	"strings"
)

type bodyOriginOptions struct {
	flags         map[string]string
	file, format  string
	control, list bool
}

func parseBodyOriginOptions(args []string) (bodyOriginOptions, error) {
	options := bodyOriginOptions{flags: map[string]string{"--activity": "", "--inputs-for": "", "--inputs-shape": "", "--field": "", "--fill-candidate": "", "--format": ""}}
	for i := 0; i < len(args); i++ {
		old, known := options.flags[args[i]]
		value := i+1 < len(args) && strings.TrimSpace(args[i+1]) != "" && !strings.HasPrefix(args[i+1], "-")
		if args[i] == "--include-control" && !options.control {
			options.control = true
		} else if args[i] == "--fill-candidates" && !options.list {
			options.list = true
		} else if known && old == "" && value {
			options.flags[args[i]] = args[i+1]
			i++
		} else if options.file == "" && !strings.HasPrefix(args[i], "-") {
			options.file = args[i]
		} else {
			return options, fmt.Errorf(bodyOriginsUsage)
		}
	}
	options.format = options.flags["--format"]
	if options.format == "" {
		options.format = "json"
	}
	return options, validateBodyOriginOptions(options)
}

func validateBodyOriginOptions(options bodyOriginOptions) error {
	flags := options.flags
	if options.file == "" || flags["--activity"] == "" || (options.control || flags["--field"] != "") && flags["--inputs-for"] == "" {
		return fmt.Errorf(bodyOriginsUsage)
	}
	if flags["--inputs-shape"] != "" && (flags["--inputs-for"] == "" || options.list || flags["--fill-candidate"] != "") {
		return fmt.Errorf("--inputs-shape requires a typed source input export with --inputs-for")
	}
	if shape := flags["--inputs-shape"]; shape != "" && shape != "ancestry-columns" && shape != "source-control-columns" && shape != "input-declaration-columns" && shape != "named-input-declaration-columns" {
		return fmt.Errorf("unknown --inputs-shape %q", shape)
	}
	if (flags["--inputs-shape"] == "input-declaration-columns" || flags["--inputs-shape"] == "named-input-declaration-columns") && (flags["--field"] != "" || options.control) {
		return fmt.Errorf("input-declaration-columns describes signature positions; --field and --include-control select graph ancestry")
	}
	if flags["--inputs-shape"] == "source-control-columns" && flags["--field"] != "" {
		return fmt.Errorf("source-control-columns contains the retained activity graph; --field selects ancestry inputs")
	}
	if options.format != "json" && options.format != "text" || options.format == "text" && (options.list || flags["--fill-candidate"] != "" || flags["--inputs-for"] != "") {
		return fmt.Errorf("--format text displays typed source origins; use --format json for candidate or input exports")
	}
	if options.list && (flags["--fill-candidate"] != "" || flags["--inputs-for"] != "" || flags["--field"] != "" || options.control) {
		return fmt.Errorf("--fill-candidates lists source assignments; use --fill-candidate to inspect or export one assignment")
	}
	if flags["--fill-candidate"] != "" && (options.control || flags["--field"] != "") {
		return fmt.Errorf("--fill-candidate describes declaration/use locations; --field and --include-control require typed value origins")
	}
	return nil
}
