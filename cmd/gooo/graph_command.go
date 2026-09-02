package main

import (
	"fmt"
	"io"
)

func runGraph(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "resolve-activity" {
		return runGraphActivityResolution(args[1:], reader, parser, stdout, stderr)
	}
	if len(args) != 2 || args[0] != "dump" {
		fmt.Fprintln(stderr, "usage: gooo graph dump <file.gooo>")
		return exitUsage
	}
	return runInspect(args[1:], reader, parser, stdout, stderr)
}
