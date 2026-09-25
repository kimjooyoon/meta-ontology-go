package main

import "errors"

const claimDependencyUsage = "usage: gooo claim dependencies <file.gooo> --json"

type claimDependencyOptions struct {
	filename string
}

func parseClaimDependencyArguments(args []string) (claimDependencyOptions, error) {
	if len(args) != 3 || args[0] != "dependencies" || args[1] == "" || args[2] != "--json" {
		return claimDependencyOptions{}, errors.New(claimDependencyUsage)
	}
	return claimDependencyOptions{filename: args[1]}, nil
}
