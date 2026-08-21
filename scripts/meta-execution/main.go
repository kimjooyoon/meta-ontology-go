package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configuration := options{}
	flag.StringVar(
		&configuration.planPath,
		"plan",
		"",
		"path to an exact self-improvement generation plan",
	)
	flag.StringVar(
		&configuration.outputPath,
		"output",
		"",
		"path for the deterministic execution manifest",
	)
	flag.Parse()
	if err := run(configuration); err != nil {
		fmt.Fprintln(os.Stderr, "meta-execution:", err)
		os.Exit(1)
	}
}
