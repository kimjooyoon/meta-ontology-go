package main

import (
	"errors"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"strings"
)

const graphActivityResolutionUsage = "usage: gooo graph resolve-activity <file.gooo> [--namespace <namespace>] [--name <name>] [--id-prefix <prefix>]"

func parseGraphActivityResolutionArguments(args []string) (string, queryengine.ActivitySelector, error) {
	if len(args) < 3 || args[0] == "" || strings.HasPrefix(args[0], "-") {
		return "", queryengine.ActivitySelector{}, errors.New(graphActivityResolutionUsage)
	}
	filename := args[0]
	selector := queryengine.ActivitySelector{}
	seen := make(map[string]bool, 3)
	for index := 1; index < len(args); index += 2 {
		if index+1 >= len(args) || seen[args[index]] || args[index+1] == "" {
			return "", queryengine.ActivitySelector{}, errors.New(graphActivityResolutionUsage)
		}
		seen[args[index]] = true
		switch args[index] {
		case "--namespace":
			selector.Namespace = args[index+1]
		case "--name":
			selector.Name = args[index+1]
		case "--id-prefix":
			selector.IDPrefix = args[index+1]
		default:
			return "", queryengine.ActivitySelector{}, errors.New(graphActivityResolutionUsage)
		}
	}
	return filename, selector, nil
}
