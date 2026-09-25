package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	settings := config{}
	flag.StringVar(&settings.root, "root", ".", "Git identity root")
	flag.StringVar(&settings.physical, "physical-root", ".", "physical repository root")
	flag.StringVar(&settings.work, "work", "", "new work directory outside the repository")
	flag.StringVar(&settings.expectedSHA, "expected-sha", "", "exact checked-out Git SHA")
	flag.StringVar(&settings.index, "git-index", "", "logical Git index output")
	flag.Parse()
	if err := execute(settings); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
