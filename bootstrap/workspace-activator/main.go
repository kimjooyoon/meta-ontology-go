package main

import (
	"flag"
	"log"
)

func main() {
	input := activationConfig{}
	flag.StringVar(&input.root, "root", "", "physical root to activate")
	flag.StringVar(&input.logical, "logical-root", "", "materialized logical root")
	flag.StringVar(&input.storage, "storage-root", "", "preserved physical root")
	flag.StringVar(&input.gitDir, "git-dir", "", "Git metadata directory")
	flag.StringVar(&input.gitIndex, "git-index", "", "logical Git index")
	flag.StringVar(&input.expectedSHA, "expected-sha", "", "exact commit identity")
	flag.StringVar(&input.materialization, "materialization", "", "materialization evidence")
	flag.StringVar(&input.evidence, "evidence", "", "activation evidence output")
	flag.Parse()
	if err := execute(input); err != nil {
		log.Fatal(err)
	}
}
