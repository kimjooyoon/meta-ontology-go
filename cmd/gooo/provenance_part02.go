package main

import (
	"errors"
	"strings"
)

func parseProvenancePublishArguments(args []string) (provenancePublishOptions, error) {
	if len(args) == 0 {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	if args[0] == "publish" {
		args = args[1:]
	}
	if len(args) == 0 {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	options := provenancePublishOptions{}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--store":
			value, next, ok := nextProvenanceValue(args, index)
			if !ok {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.store = value
			index = next
		case "--evidence", "--input":
			value, next, ok := nextProvenanceValue(args, index)
			if !ok {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.evidence = value
			index = next
		default:
			if strings.HasPrefix(arg, "-") || options.source != "" {
				return provenancePublishOptions{}, errors.New(provenancePublishUsage)
			}
			options.source = arg
		}
	}
	if options.source == "" || options.store == "" || options.evidence == "" {
		return provenancePublishOptions{}, errors.New(provenancePublishUsage)
	}
	return options, nil
}
func nextProvenanceValue(args []string, index int) (string, int, bool) {
	if index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
		return "", index, false
	}
	return args[index+1], index + 1, true
}
