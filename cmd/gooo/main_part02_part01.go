package main

import (
	"errors"
	"strings"
)

func parseAnalyzeDeltaArguments(args []string) (analyzeDeltaOptions, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
	}
	o := analyzeDeltaOptions{authority: args[0]}
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--go", "--generated-go", "--input":
			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
			}
			o.goFiles = append(o.goFiles, args[i+1])
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
			}
			o.goFiles = append(o.goFiles, arg)
		}
	}
	if len(o.goFiles) == 0 {
		return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
	}
	return o, nil
}
