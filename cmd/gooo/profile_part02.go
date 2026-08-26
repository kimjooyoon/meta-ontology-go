package main

import (
	"errors"
	"strconv"
	"strings"
)

type profileOptions struct {
	filename string
	entry    string
	samples  int
}

func parseProfileArguments(args []string) (profileOptions, error) {
	options := profileOptions{samples: 5}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--entry":
			if options.entry != "" || index+1 >= len(args) {
				return profileOptions{}, errors.New(profileUsage)
			}
			index++
			options.entry = args[index]
		case "--samples":
			if index+1 >= len(args) {
				return profileOptions{}, errors.New(profileUsage)
			}
			index++
			value, err := strconv.Atoi(args[index])
			if err != nil {
				return profileOptions{}, errors.New(profileUsage)
			}
			options.samples = value
		default:
			if strings.HasPrefix(args[index], "-") || options.filename != "" {
				return profileOptions{}, errors.New(profileUsage)
			}
			options.filename = args[index]
		}
	}
	if options.filename == "" || strings.TrimSpace(options.entry) == "" || options.samples < 1 || options.samples > 20 {
		return profileOptions{}, errors.New(profileUsage)
	}
	return options, nil
}
