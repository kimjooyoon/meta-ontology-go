package main

import (
	"fmt"
	"io"
)

const claimUsage = "usage: gooo claim <resolve|dependencies> [args]"

func runClaim(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, claimUsage)
		return exitUsage
	}
	switch args[0] {
	case "resolve":
		return runClaimResolution(args, reader, parser, stdout, stderr)
	case "dependencies":
		return runClaimDependencies(args, reader, parser, stdout, stderr)
	default:
		fmt.Fprintln(stderr, claimUsage)
		return exitUsage
	}
}
