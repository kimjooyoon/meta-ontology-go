package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configuration := options{}
	flag.StringVar(&configuration.planPath, "plan", "", "generation plan path")
	flag.StringVar(&configuration.executionPath, "execution", "", "execution manifest path")
	flag.StringVar(&configuration.receiptsPath, "receipts", "", "receipt report path")
	flag.StringVar(&configuration.outputPath, "output", "", "artifact provenance output path")
	flag.Parse()
	if err := run(configuration); err != nil {
		fmt.Fprintln(os.Stderr, "meta-provenance:", err)
		os.Exit(1)
	}
}
