package main

import (
	"flag"
	"log"
)

func main() {
	input := cutoverConfig{}
	flag.StringVar(&input.root, "root", "", "current Git worktree")
	flag.StringVar(&input.physical, "physical-root", "", "candidate physical tree")
	flag.StringVar(&input.authority, "authority-manifest", "", "CI authority manifest")
	flag.StringVar(&input.expectedSHA, "expected-sha", "", "exact source commit")
	flag.StringVar(&input.backup, "backup", "", "logical tree backup")
	flag.StringVar(&input.evidence, "evidence", "", "cutover evidence output")
	flag.BoolVar(&input.apply, "apply", false, "activate and stage the physical tree")
	flag.Parse()
	if err := execute(input); err != nil {
		log.Fatal(err)
	}
}
