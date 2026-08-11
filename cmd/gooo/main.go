package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: gooo <check|generate|analyze|lsp> [args]")
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "gooo: command %q is not implemented yet\n", os.Args[1])
	os.Exit(1)
}
