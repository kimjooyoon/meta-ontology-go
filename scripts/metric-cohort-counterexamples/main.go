// Command metric-cohort-counterexamples consumes one exact original producer artifact.
package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	Root, Scratch, Input, Concept, Metadata, Archive, Output string
	Origin                                                   origin
}

func main() {
	var cfg config
	flag.StringVar(&cfg.Root, "root", "", "exact subject repository")
	flag.StringVar(&cfg.Scratch, "scratch", "", "external scratch directory")
	flag.StringVar(&cfg.Input, "input", "", "unpacked producer directory")
	flag.StringVar(&cfg.Concept, "concept", "", "source-bound concept artifact")
	flag.StringVar(&cfg.Metadata, "metadata", "", "GitHub producer artifact metadata")
	flag.StringVar(&cfg.Archive, "archive", "", "exact producer zip archive")
	flag.StringVar(&cfg.Output, "output", "", "new external result directory")
	flag.StringVar(&cfg.Origin.Repository, "repository", "", "expected repository")
	flag.StringVar(&cfg.Origin.SubjectSHA, "subject-sha", "", "expected source revision")
	flag.Int64Var(&cfg.Origin.RunID, "run-id", 0, "original producer workflow run")
	flag.IntVar(&cfg.Origin.Attempt, "run-attempt", 0, "original producer attempt")
	flag.Int64Var(&cfg.Origin.ArtifactID, "artifact-id", 0, "producer job artifact ID")
	flag.StringVar(&cfg.Origin.ArtifactDigest, "artifact-digest", "", "producer job archive digest")
	flag.Parse()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
