package main

import "errors"

const claimResolutionUsage = "usage: gooo claim resolve <file.gooo> --activity <name> --json"

type claimResolutionOptions struct {
	filename string
	activity string
}

func parseClaimResolutionArguments(args []string) (claimResolutionOptions, error) {
	if len(args) != 5 || args[0] != "resolve" || args[1] == "" || args[2] != "--activity" || args[3] == "" || args[4] != "--json" {
		return claimResolutionOptions{}, errors.New(claimResolutionUsage)
	}
	return claimResolutionOptions{filename: args[1], activity: args[3]}, nil
}
