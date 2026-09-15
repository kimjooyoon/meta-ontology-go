package main

import "errors"

type options struct {
	expectedHead   string
	contract       string
	golden         string
	artifact       string
	replay         string
	unknownEmitter string
	profile        string
	out            string
}

func parseOptions(args []string) (options, error) {
	var value options
	for index := 0; index < len(args); index++ {
		if index+1 >= len(args) {
			return options{}, errors.New("every option requires a value")
		}
		next := args[index+1]
		switch args[index] {
		case "--expected-head":
			value.expectedHead = next
		case "--contract":
			value.contract = next
		case "--golden":
			value.golden = next
		case "--artifact":
			value.artifact = next
		case "--replay":
			value.replay = next
		case "--unknown-emitter":
			value.unknownEmitter = next
		case "--profile":
			value.profile = next
		case "--out":
			value.out = next
		default:
			return options{}, errors.New("unknown option: " + args[index])
		}
		index++
	}
	if value.expectedHead == "" || value.contract == "" || value.golden == "" || value.artifact == "" ||
		value.replay == "" || value.unknownEmitter == "" || value.profile == "" || value.out == "" {
		return options{}, errors.New("all experiment inputs are required")
	}
	return value, nil
}
