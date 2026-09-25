package main

import "fmt"

func parseInvokeArguments(args []string) (invokeOptions, error) {
	options := invokeOptions{}
	for len(args) != 0 {
		switch args[0] {
		case "--entry":
			if len(args) < 2 || options.entry != "" {
				return invokeOptions{}, fmt.Errorf("invalid entry option")
			}
			options.entry, args = args[1], args[2:]
		case "--input":
			if len(args) < 2 || options.input != "" {
				return invokeOptions{}, fmt.Errorf("invalid input option")
			}
			options.input, args = args[1], args[2:]
		default:
			if options.source != "" || len(args[0]) == 0 || args[0][0] == '-' {
				return invokeOptions{}, fmt.Errorf("invalid source argument")
			}
			options.source, args = args[0], args[1:]
		}
	}
	if options.entry == "" || options.input == "" || options.source == "" {
		return invokeOptions{}, fmt.Errorf("missing required argument")
	}
	return options, nil
}
