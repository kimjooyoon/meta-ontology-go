package main

import (
	"fmt"
	"strings"
)

type starterOptions struct {
	template, module, destination string
	moduleProvided                bool
}

func parseStarterOptions(args []string) (starterOptions, error) {
	options := starterOptions{template: "app", module: "boundedint"}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--template", "--module":
			if index+1 >= len(args) {
				return options, fmt.Errorf("%s", initUsage)
			}
			if args[index] == "--template" {
				options.template = args[index+1]
			} else {
				if strings.TrimSpace(args[index+1]) == "" {
					return options, fmt.Errorf("%s", initUsage)
				}
				options.module, options.moduleProvided = args[index+1], true
			}
			index++
		default:
			if strings.HasPrefix(args[index], "-") || options.destination != "" {
				return options, fmt.Errorf("%s", initUsage)
			}
			options.destination = args[index]
		}
	}
	return options, nil
}
